package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

var (
	githubOauthConfig *oauth2.Config
	store             *sessions.CookieStore
)

type GitHubUser struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

func init() {
	githubOauthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		RedirectURL:  "http://localhost:8080/callback",
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
	}

	store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))
}

func main() {
	if githubOauthConfig.ClientID == "" || githubOauthConfig.ClientSecret == "" {
		log.Fatal("GitHub OAuth credentials not set. Please set GITHUB_CLIENT_ID and GITHUB_CLIENT_SECRET environment variables.")
	}

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/callback", callbackHandler)
	http.HandleFunc("/profile", profileHandler)
	http.HandleFunc("/logout", logoutHandler)

	fmt.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>GitHub OAuth Example</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 600px; margin: 50px auto; padding: 20px; }
        .btn { display: inline-block; padding: 10px 20px; background: #333; color: white; text-decoration: none; border-radius: 5px; }
        .btn:hover { background: #555; }
    </style>
</head>
<body>
    <h1>GitHub OAuth Login Example</h1>
    {{if .User}}
        <p>Welcome back, {{.User}}!</p>
        <a href="/profile" class="btn">View Profile</a>
        <a href="/logout" class="btn">Logout</a>
    {{else}}
        <p>Please log in with your GitHub account to continue.</p>
        <a href="/login" class="btn">Login with GitHub</a>
    {{end}}
</body>
</html>`

	t := template.Must(template.New("home").Parse(tmpl))
	
	data := struct {
		User string
	}{
		User: getStringFromSession(session, "user"),
	}
	
	t.Execute(w, data)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	url := githubOauthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	
	token, err := githubOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}

	client := githubOauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var user GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		http.Error(w, "Failed to decode user info", http.StatusInternalServerError)
		return
	}

	session, _ := store.Get(r, "session")
	session.Values["user"] = user.Login
	session.Values["name"] = user.Name
	session.Values["email"] = user.Email
	session.Values["avatar_url"] = user.AvatarURL
	session.Save(r, w)

	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	
	user := getStringFromSession(session, "user")
	if user == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>Profile - GitHub OAuth Example</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 600px; margin: 50px auto; padding: 20px; }
        .profile { background: #f5f5f5; padding: 20px; border-radius: 10px; }
        .avatar { border-radius: 50%; width: 100px; height: 100px; }
        .btn { display: inline-block; padding: 10px 20px; background: #333; color: white; text-decoration: none; border-radius: 5px; margin-top: 20px; }
        .btn:hover { background: #555; }
    </style>
</head>
<body>
    <h1>Your GitHub Profile</h1>
    <div class="profile">
        {{if .AvatarURL}}<img src="{{.AvatarURL}}" alt="Avatar" class="avatar"><br><br>{{end}}
        <strong>Username:</strong> {{.User}}<br>
        {{if .Name}}<strong>Name:</strong> {{.Name}}<br>{{end}}
        {{if .Email}}<strong>Email:</strong> {{.Email}}<br>{{end}}
    </div>
    <a href="/" class="btn">Home</a>
    <a href="/logout" class="btn">Logout</a>
</body>
</html>`

	t := template.Must(template.New("profile").Parse(tmpl))
	
	data := struct {
		User      string
		Name      string
		Email     string
		AvatarURL string
	}{
		User:      user,
		Name:      getStringFromSession(session, "name"),
		Email:     getStringFromSession(session, "email"),
		AvatarURL: getStringFromSession(session, "avatar_url"),
	}
	
	t.Execute(w, data)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	session.Values = make(map[interface{}]interface{})
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func getStringFromSession(session *sessions.Session, key string) string {
	if val, ok := session.Values[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}