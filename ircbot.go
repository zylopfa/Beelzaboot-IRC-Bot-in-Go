package main

import (
  "net"
  "io/ioutil"
  "os"
  "fmt"
  "time"
  "strings"
  "math/rand"
)


func main() {

  if len(os.Args) != 3 {
   fmt.Fprintf(os.Stderr,"Usage: %s host:port channel \n",os.Args[0])
   os.Exit(1)
  }

  rand.Seed(time.Now().UnixNano())

  channel := os.Args[2]
  service := os.Args[1]
 
  nick := "Beelzaboot"

  tcpAddr,err := net.ResolveTCPAddr("tcp4",service)
  checkError(err)

  conn,err := net.DialTCP("tcp",nil,tcpAddr)
  checkError(err)

  timeoutDuration := 600 * time.Second

  p := make([]byte,1)

  var line string
  var strLen int = 0

  var identified bool = false
  var haveJoinedChannel = false

  for {

    conn.SetReadDeadline(time.Now().Add(timeoutDuration))
    _,err := conn.Read(p)

    checkError(err)

    if p[0] != 0x0a && strLen < 512 {
      line += string(p)
      strLen++
      continue
    }
    fmt.Printf("%s\n",line)


    if ( strings.Contains(line,"NOTICE") && !identified ) {
      fmt.Print("Identifying!\n")

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


      fmt.Printf("From         : %s\n",from)
      fmt.Printf("From domain  : %s\n",domain)
      fmt.Printf("To           : %s\n",to)
      fmt.Printf("Command      : %s\n",cmd)
      fmt.Printf("Message: %s\n",msgPart)
      fmt.Print("\n")


      // Message directed to the BOT
      if  cmd == "PRIVMSG" && to == nick   {
        fmt.Printf("Message to the bot!\n")

        if strings.Contains(msgPart,"\x01VERSION\x01") {
          fmt.Printf("%s want my version!\n",from)
          _,err = conn.Write([]byte("NOTICE " + from + " :" + "\x01VERSION Belzaboot 1.0 - The Canadian Devil IRC Bot\x01\r\n" ))
         checkError(err)
        }

      // End of bot directed messages
      } else {
        if strings.Contains(msgPart,"\x01ACTION") {
          actionList := strings.Split(msgPart,"\x01")
          if  len(actionList) == 3 {
            actionList[1] = "Here " + randomPronoun() + " have some canadough!!"
            _,err = conn.Write([]byte(cmd + " " + from + " :" +  strings.Join(actionList,"\x01")))

            fmt.Printf("We are actioned!\n")
          }
        }

        if strings.Contains(msgPart,nick) && strings.Contains(to,"#") {
           fmt.Printf("Addressing by name " + to);

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

  result,err := ioutil.ReadAll(conn)
  checkError(err)

  fmt.Println(string(result))

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
