package main

import (
	"log"
	"os"

	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/api/handlers"
	middlewares "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/api/middleware"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables. Uncomment when running locally and not in container
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	addr := os.Getenv("DEV_ADDRESS")
	secret := os.Getenv("SECRET")

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	app := &handlers.Application{
		ErrorLog: errorLog,
		InfoLog:  infoLog,
	}

	//gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	store := cookie.NewStore([]byte(secret))
	store.Options(sessions.Options{
		MaxAge:   12 * 60 * 60, // 12 hours
		HttpOnly: true,
		Secure:   true, // true in production
	})
	router.Use(sessions.Sessions("mysession", store))

	router.Use(middlewares.SecureHeaders())
	router.Use(middlewares.CorsMiddleware())

	apiV1 := router.Group("/api/v1")
	{
		//apiV1.POST("/user/signup", app.SignUpUser)
		//apiV1.POST("/user/login", app.LoginUser)
		//apiV1.POST("/user/logout", app.LogoutUser)
		apiV1.GET("/search/word", app.DefineWord)
		apiV1.PUT("/search/phrase", app.DefinePhrase)

		/* Uncomment for when we add premium features
		authorized := apiV1.Group("/")
		authorized.Use(middlewares.AuthRequired())
		{
			authorized.GET("/search/word", app.DefineWord)
			authorized.PUT("/search/phrase", app.DefinePhrase)
		}
		*/

	}
	infoLog.Printf("Starting server on %s", addr)

	//router.RunTLS(addr, "./tls/cert.pem", "./tls/key.pem") TO-DO: Server over HTTPS when figure out how to get certificates
	router.Run(addr)
	if err != nil {
		errorLog.Fatal(err)
	}
}
