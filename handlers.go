package main

import (
	"bytes"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
)

// IndexHandler shows main page of the site and does not require authentication
func IndexHandler(w http.ResponseWriter, r *http.Request) {

	data := make(map[string]interface{})

	cookie, err := r.Cookie("sid")

	if err == nil {
		session := sessionStore.Get(cookie.Value)
		if session.isNew == false {
			data["user"] = session.value["user"]
		}
	}

	renderTemplate(w, "index.html", data)
}

// VideoHandler shows live video from uav
func VideoHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("sid")

	if err != nil {
		http.Redirect(w, r, "/login", 302)
		return
	}

	session := sessionStore.Get(cookie.Value)
	if session.isNew {
		http.Redirect(w, r, "/login", 302)
		return
	}

	session.Update()

	data := make(map[string]interface{})
	data["user"] = session.value["user"]

	renderTemplate(w, "video.html", data)
}

// SnapshotHandler show a page with snapshot
func SnapshotHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("sid")
	//fmt.Println(cookie)
	//fmt.Println(cookie.Value)

	if err != nil {
		http.Redirect(w, r, "/login", 302)
		return
	}

	session := sessionStore.Get(cookie.Value)

	if session.isNew {
		http.Redirect(w, r, "/login", 302)
		return
	}

	// update timeAccessed
	session.Update()

	data := make(map[string]interface{})
	data["user"] = session.value["user"]

	renderTemplate(w, "snapshot.html", data)
}

// LoginHandler handles user login
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]interface{})

	if r.Method == "GET" {
		renderTemplate(w, "login.html", data)
	}

	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")

		loggedin := false

		if user, ok := FindUser(username, password); ok {
			fmt.Println(user)
			session := NewSession(map[interface{}]interface{}{"user": user})
			//fmt.Println(session)
			session.isNew = false
			sessionStore.Set(session.sid, session)
			cookie := &http.Cookie{Name: "sid", Value: session.sid}

			http.SetCookie(w, cookie)

			loggedin = true

			http.Redirect(w, r, "/", 302)
		}

		data["username"] = username
		data["errorMessage"] = "Kullanıcı adı veya Şifresi hatalı"

		if !loggedin {
			renderTemplate(w, "login.html", data)
		}
	}
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("sid")

	if err == nil {
		// Delete session for given session id (sid)
		sessionStore.Delete(cookie.Value)
		cookie := &http.Cookie{Name: "sid", Value: "", MaxAge: -1}
		http.SetCookie(w, cookie)

		http.Redirect(w, r, "/", 302)
		return
	}

	http.Redirect(w, r, "/", 302)
}

func JpegHandler(w http.ResponseWriter, r *http.Request) {

	cookie, err := r.Cookie("sid")

	if err != nil {
		http.Redirect(w, r, "/login", 302)
		return
	}

	session := sessionStore.Get(cookie.Value)
	if session.isNew {
		http.Redirect(w, r, "/login", 302)
		return
	}

	session.Update()

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("cache-control", "priviate, max-age=0, no-cache")
	w.Header().Set("pragma", "no-cache")
	w.Header().Set("expires", "-1")

	w.Write(jpegImage)
}

func MotionJpegHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("sid")
	//fmt.Println(cookie)
	//fmt.Println(cookie.Value)

	if err != nil {
		http.Redirect(w, r, "/login", 302)
		return
	}

	session := sessionStore.Get(cookie.Value)
	if session.isNew {
		http.Redirect(w, r, "/login", 302)
		return
	}

	session.Update()

	log.Println("Serve streaming")
	m := multipart.NewWriter(w)
	w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary="+m.Boundary())
	w.Header().Set("Connection", "close")
	w.Header().Set("cache-control", "priviate, max-age=0, no-cache")
	w.Header().Set("pragma", "no-cache")
	w.Header().Set("expires", "-1")

	header := textproto.MIMEHeader{}
	var buf bytes.Buffer
	for {
		mutex.RLock()
		buf.Reset()

		_, err := buf.Write(jpegImage)
		//err = jpeg.Encode(&buf, img, nil)
		//w.Write(jpegImage)
		mutex.RUnlock()
		if err != nil {
			break
		}
		header.Set("Content-Type", "image/jpeg")
		header.Set("Content-Length", fmt.Sprint(buf.Len()))
		mw, err := m.CreatePart(header)
		if err != nil {
			break
		}
		mw.Write(buf.Bytes())
		if flusher, ok := mw.(http.Flusher); ok {
			flusher.Flush()
		}
	}
	log.Println("Stop streaming")
}
