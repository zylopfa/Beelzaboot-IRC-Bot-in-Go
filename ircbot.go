package main

import (
  "net"
  "os"
  "fmt"
  "time"
  "strings"
  "math/rand"
  "log"
  "database/sql" 
  _ "github.com/mattn/go-sqlite3" 
)



type ircmessage struct {
    from string
    domain string
    to string
    cmd string
    msg string
}


func main() {

  start := time.Now()
  var jointime time.Duration

  debug := 1  // 0: dont print anything, but fatal errors, 1: print only join time, 2: print all info

  if len(os.Args) != 5 {
   fmt.Fprintf(os.Stderr,"Usage: %s host:port channel nick bindaddress\n",os.Args[0])
   os.Exit(1)
  }

  rand.Seed(time.Now().UnixNano())

  channel := os.Args[2]
  service := os.Args[1]
 
  nick := os.Args[3]

  bindAddr,err := net.ResolveTCPAddr("tcp", os.Args[4] + ":0") 

  tcpAddr,err := net.ResolveTCPAddr("tcp6",service)
  checkError(err)

  conn,err := net.DialTCP("tcp",bindAddr,tcpAddr)
  checkError(err)

  timeoutDuration := 600 * time.Second

  p := make([]byte,1)

  var line string
  var strLen int = 0

  var identified bool = false
  var haveJoinedChannel = false

  // sqlite3 database initialization
  db :=  prepareSqlite(nick)

  for {

    conn.SetReadDeadline(time.Now().Add(timeoutDuration))
    _,err := conn.Read(p)

    checkError(err)

    if p[0] != 0x0a && strLen < 512 {
      line += string(p)
      strLen++
      continue
    }

    if ( debug == 2 ) {  fmt.Printf("%s\n",line) }


    if ( strings.Contains(line,"NOTICE") && !identified ) {
      _,err = conn.Write([]byte("USER " + nick + " \"\" \"\" :" + nick + "\r\n"))
      checkError(err)

      _,err = conn.Write([]byte("NICK " + nick + "\r\n"))
      checkError(err)
      identified = true
    }


    if ( strings.Contains(line,"PING") ) {
      _,err = conn.Write([]byte(strings.Replace(line,"I","O",-1)))
      checkError(err)

      if ( !haveJoinedChannel ) {
        _,err = conn.Write([]byte("join " + channel + "\r\n" ))
       checkError(err)
       haveJoinedChannel = true
      }
    }

    s := strings.Split(line,":")


    if ( haveJoinedChannel && len(s) >= 2 ) {

      if ( jointime == 0 ) {
        jointime = time.Since(start)
        if debug >= 1 {log.Printf("Time taken to join channel %s\n", jointime) }
      }

      var infoPart string = s[1]
      var msgPart string = ""

      if ( len(s) >= 3 ) {
        msgPart = strings.Join(s[2:],":")
      }

      infoArr := strings.Split(infoPart," ")

      userArr := strings.Split(infoArr[0],"!")
      from := ""
      domain := ""
      cmd := ""
      to := ""

      if ( len(userArr) >= 2 ) {
        from = userArr[0]
        domain = userArr[1]
      }

      if ( len(infoArr) >= 3 ) {
        cmd = infoArr[1]
        to  = infoArr[2]
      }

      if debug >= 2 {
        fmt.Printf("From         : %s\n",from)
        fmt.Printf("From domain  : %s\n",domain)
        fmt.Printf("To           : %s\n",to)
        fmt.Printf("Command      : %s\n",cmd)
        fmt.Printf("Message: %s\n",msgPart)
        fmt.Print("\n")
      }

      irc := ircmessage{from: from, domain: domain,to: to,cmd:cmd,msg:msgPart}

      if ( irc.from != "" ) {
        db_logIrcMessage(db,irc)
      }


      // Message directed to the BOT
      if  cmd == "PRIVMSG" && to == nick   {
        if debug >= 2 {
          fmt.Printf("Message to the bot!\n")
        }

        if strings.Contains(msgPart,"\x01VERSION\x01") {
          if debug >= 2 {
            fmt.Printf("%s want my version!\n",from)
          }
          _,err = conn.Write([]byte("NOTICE " + from + " :" + "\x01VERSION Belzaboot 1.0 - The Canadian Devil IRC Bot\x01\r\n" ))
         checkError(err)
        }

      // End of bot directed messages
      } else {

        if strings.Contains(msgPart,nick) && strings.Contains(to,"#") {

           if debug >= 2 {
             fmt.Printf("Addressing by name " + to);
           }
           randString := randomAnswer();

           if strings.Contains(randString,"%s") {
              _,err = conn.Write([]byte("PRIVMSG" + " " + to + " :" +  fmt.Sprintf(randString,from) + "\r\n" ))     
           } else {
              _,err = conn.Write([]byte("PRIVMSG" + " " + to + " :" +  randString + "\r\n" ))     

           }
        }
      }

    }

    strLen = 0
    line = ""

  }


  os.Exit(0)
}

func checkError(err error) {
  if err != nil {
    fmt.Fprintf(os.Stderr,"Fatal error: %s\n", err.Error())
    os.Exit(1)
  }
}


func randomPronoun() (string) {

  pronouns := []string{
    "guy",
    "friend",
    "bro",
    "mate",
  }
  return pronouns[rand.Int() % len(pronouns)]
}

func randomAnswer() (string) {
  answers := []string{
    "if you buy some canadough you can help Terrance and Phillip save canada!",
    "You're a boner biting bastard, uncle fucka!",
    "You discovered my plan, but too late! [lets out a flaming turd] Now the souls of all Canadians belong to me!",
    "Heheheh! Hahaha! [lets out a flaming turd]",
    "Well well, my overachieving doppelganger! You're no match for Canadian Satan! [lets out a flaming turd]",
    "I Beelzaboot is a fremium IRC Bot!! Free\"mium\". The \"mium\" is Latin for \"not really.\"",
    "%s just because your dad make a good living with my music doesn't mean you can go blow it all on Canadough!",
  }
  return answers[rand.Int() % len(answers)];
}


func prepareSqlite(nick string) (*sql.DB) {

  if _, err := os.Stat("./db"); os.IsNotExist(err) {
    os.Mkdir("./db", 0600)
  }

  dbFile := "./db/" + nick + ".db"

  db, err := sql.Open("sqlite3",dbFile)

  if err != nil {
    log.Fatal(err)
  }

  if _, err := os.Stat(dbFile ); os.IsNotExist(err) {

    fmt.Printf("File '%s' doesn't exist, creating db ..\n",dbFile)

    sqlStatement := `
      CREATE TABLE sqlLog (fromNick TEXT, fromDomain TEXT, toEntity TEXT, command TEXT,message TEXT, msgDate DATETIME)
    `

    _, err = db.Exec(sqlStatement)
    if err != nil {
      log.Printf("%q: %s\n", err, sqlStatement)
      return nil
    }
  }

  return db

}




func db_logIrcMessage(db *sql.DB,irc ircmessage) {

  tx, err := db.Begin()

  if err != nil {
    log.Fatal(err)
  }

  stmt, err := tx.Prepare("insert into sqlLog(fromNick, fromDomain, toEntity, command, message, msgDate ) values(?,?,?,?,?,datetime('now'))")
  if err != nil {
    log.Fatal(err)
  }

  defer stmt.Close()

  _,err = stmt.Exec(irc.from, irc.domain, irc.to, irc.cmd, irc.msg)

  if err != nil {
    log.Fatal(err)
  }

  tx.Commit()

}
