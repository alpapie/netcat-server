package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"net-cat/lib"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Profil struct {
	Name string
	Addr string
	Con  net.Conn
}
type User struct {
	Profil          Profil
	NewMesssage     string
	DeconnectClient bool
}

var Profils = []Profil{}
var PORT = ""

func main() {
	err := ioutil.WriteFile("log/log.txt", []byte(""), 0644)
	lib.Errorstr(err, "")

	tab := os.Args[1:]

	if len(tab) > 0 {
		t, err := strconv.Atoi(tab[0])
		if err != nil {
			lib.Errorstr(nil, "Only number")
		}
		if t >= 1024 && t <= 65535 {
			PORT = ":" + tab[0]
		} else {
			lib.Errorstr(nil, "The port must be between 1024 and 65535")
		}
	} else {
		PORT = ":8989"
	}

	file, er := ioutil.ReadFile("Welcome.txt")
	lib.Errorstr(er, "")
	Listen, err := net.Listen("tcp", PORT)
	lib.Errorstr(err, "")
	fmt.Println("server is run on port " + PORT)
	getCon(Listen, file)

}

//*****************************************************GET THE INCOMMING  CONNECTION *****************************
func getCon(Listen net.Listener, file []byte) {
	defer Listen.Close()

	// var NewUser map[string]string
	var m sync.Mutex
	var NewUser = make(chan User)
	for {
		con, err := Listen.Accept()
		fmt.Println(len(Profils))
		if len(Profils) < 10 {
			lib.Errorstr(err, "")
			con.Write(file)
			go handleIncomingRequest(con, NewUser, m)
			go Register(NewUser, m)
		} else {
			con.Write([]byte("Client list is ful try a gain later"))
			con.Close()
		}
	}

}

//***************************************************** REGISTER  NEW USER*****************************
func Register(NewUser chan User, m sync.Mutex) {

	for ch := range NewUser {
		if ch.Profil.Name != "" {
			for _, v := range Profils {
				v.Con.Write([]byte("\n" + strings.ReplaceAll(ch.Profil.Name, "\n", "") + " has joined our chat..."))
				v.Con.Write([]byte("\n[" + time.Now().Format("2006-01-02 15:04:05") + "][" + strings.ReplaceAll(v.Name, "\n", "") + "]:"))
			}
			Profils = append(Profils, ch.Profil)
		} else if ch.NewMesssage != "" {
			m.Lock()
			file, err := os.OpenFile("log/log.txt", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
			defer file.Close()
			lib.Errorstr(err, "")
			_, eer := file.WriteString(ch.NewMesssage)
			lib.Errorstr(eer, "")

			m.Unlock()
			for _, v := range Profils {
				if ch.Profil.Addr != v.Addr {
					v.Con.Write([]byte("\r\033K[" + strings.ReplaceAll(ch.NewMesssage, "\n", "")))
					v.Con.Write([]byte("\n[" + time.Now().Format("2006-01-02 15:04:05") + "][" + strings.ReplaceAll(v.Name, "\n", "") + "]:"))
				}
			}
		}
		if ch.DeconnectClient {
			name := ""
			for i, v := range Profils {
				if ch.Profil.Addr == v.Addr {
					name = v.Name
					ch.Profil.Con.Close()
					Profils = append(Profils[:i], Profils[i+1:]...)
					break
				}
			}
			for _, v := range Profils {
				v.Con.Write([]byte("\n" + strings.ReplaceAll(name, "\n", "") + " has left our chat..."))
				v.Con.Write([]byte("\n[" + time.Now().Format("2006-01-02 15:04:05") + "][" + strings.ReplaceAll(v.Name, "\n", "") + "]:"))
			}
		}
	}

}

//*************************************************** TRAITEMENT FOR THE CONNECT *****************************
func handleIncomingRequest(conn net.Conn, NewUser chan User, m sync.Mutex) {
	IsFirthMessg := true
	user := User{}
	user.Profil.Addr = conn.RemoteAddr().String()
	user.Profil.Con = conn
	name := ""
	headerMessage := ""

	for {
		var str = ""
		user.NewMesssage = ""
		if !IsFirthMessg {
			headerMessage = "[" + time.Now().Format("2006-01-02 15:04:05") + "][" + strings.ReplaceAll(name, "\n", "") + "]:"
			conn.Write([]byte(headerMessage))
		}
		buffer := make([]byte, 1024)
		_, err := conn.Read(buffer)
		str = lib.GetString(buffer)
		if err == nil {
			if IsFirthMessg {
				IsFirthMessg = false
				user.Profil.Name = str
				name = str
				NewUser <- user
				m.Lock()
				data, err := ioutil.ReadFile("log/log.txt")
				lib.Errorstr(err, "")
				conn.Write(data)
				m.Unlock()
			} else {
				user.Profil.Name = ""
				if str != "\n" {
					user.NewMesssage = headerMessage + str
				}
				NewUser <- user
			}
		}else {
			user.DeconnectClient = true
			user.NewMesssage = str
			NewUser <- user
			break
		}
	}
}
