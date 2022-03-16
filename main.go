package main

import (
	"fmt"
	"go-test-fiber/model"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	jwtware "github.com/gofiber/jwt/v2"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

var db *sqlx.DB

const jwtSecret = "infinitas"

func main() {
	var err error
	db, err = sqlx.Open("mysql", "root:admin@tcp(db)/techcoach")
	if err != nil {
		panic(fmt.Sprintf("sdsd %d",err))
	}
	app := fiber.New()
	app.Use(logger.New(logger.Config{TimeZone: "Asia/Bangkok"}))
	// middleware
	app.Use("/hello", jwtware.New(jwtware.Config{
		SigningMethod: "HS256",
		SigningKey:    []byte(jwtSecret),
		SuccessHandler: func(ctx *fiber.Ctx) error {
			return ctx.Next()
		},
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			return fiber.ErrUnauthorized
		},
	}))
	app.Use(cors.New(cors.Config{
		AllowHeaders: "*",
		AllowMethods: "*",
		AllowOrigins: "*",
	}))

	app.Get("/", helloWorld)
	app.Post("/signup", Signup)
	app.Post("/login", Login)
	app.Get("/hello", Hello)

	app.Listen(":8000")
}
func helloWorld(ctx *fiber.Ctx) error {
	return ctx.SendString("Hello go")
}
func Signup(ctx *fiber.Ctx) error {
	request := model.SignupRequest{}
	err := ctx.BodyParser(&request)
	if err != nil {
		return err
	}
	if request.Username == "" || request.Password == "" {
		return fiber.ErrUnprocessableEntity
	}
	//  แปลง pass bcrypt
	password, err := bcrypt.GenerateFromPassword([]byte(request.Password), 10)
	if err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "test1")
	}
	query := "insert techcoach (username , password) values (?,?)"
	res, err := db.Exec(query, request.Username, string(password)) // แปลง string(password) กลับ
	if err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, fmt.Sprintln("test3" + err.Error()))
	}

	id, err := res.LastInsertId()
	if err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "test3")
	}
	user := model.User{
		Id:       int(id),
		Username: request.Username,
		Password: string(password),
	}
	return ctx.Status(fiber.StatusCreated).JSON(user)
}
func Login(ctx *fiber.Ctx) error {
	request := model.LoginRequest{}
	err := ctx.BodyParser(&request)
	if err != nil {
		return err
	}

	if request.Username == "" || request.Password == "" {
		return fiber.ErrUnprocessableEntity
	}

	user := model.User{}
	query := "select id , username , password from techcoach where username=?"

	err = db.Get(&user, query, request.Username)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Incorrect username or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password))
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Incorrect username or password")
	}

	claims := jwt.StandardClaims{
		Issuer:    strconv.Itoa(user.Id),
		ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	token, err := jwtToken.SignedString([]byte(jwtSecret)) // ใส่ secret key เข้าไป และแปลง jwt เป็น slice of byte
	if err != nil {
		return fiber.ErrInternalServerError
	}

	return ctx.JSON(fiber.Map{
		"jwtToken": token,
		"status" : "success",
	})

	return fiber.NewError(fiber.StatusOK, "login success !")
}
func Hello(ctx *fiber.Ctx) error {
	return ctx.SendString("Hello eiei ~!")
}

func Fiber() {
	app := fiber.New(fiber.Config{
		Prefork: true, // Prefork คือ option เพื่อรับงานหลาย process
	})
	// Middleware
	app.Use("/hello", func(ctx *fiber.Ctx) error {
		ctx.Locals("name", "mick") // ค่าจาก middleware ส่งต่อค่าไปยัง handler
		err := ctx.Next()
		return err
	})
	app.Use(requestid.New())

	app.Use(cors.New(cors.Config{
		AllowHeaders: "*",
		AllowMethods: "*",
		AllowOrigins: "*",
	}))

	app.Use(logger.New(logger.Config{TimeZone: "Asia/Bangkok"}))
	// GET
	app.Get("/hello", func(ctx *fiber.Ctx) error {
		name := ctx.Locals("name")
		return ctx.SendString(fmt.Sprintf("Get Hello %v", name))
	})
	// Post
	app.Post("/hello", func(ctx *fiber.Ctx) error {
		return ctx.SendString("Post hello !")
	})
	// params and optional
	app.Get("/hello/:name/:surname", func(ctx *fiber.Ctx) error {
		name := ctx.Params("name")
		surname := ctx.Params("surname")
		return ctx.SendString("Name : " + name + " Surname : " + surname)
	})
	// Params int
	app.Get("/hello/:id", func(ctx *fiber.Ctx) error {
		id, err := ctx.ParamsInt("id")
		if err != nil {
			return fiber.ErrBadRequest
		}
		return ctx.SendString(fmt.Sprintf("Hello you id : %v", id))
	})
	// Query
	app.Get("/query", func(ctx *fiber.Ctx) error {
		name := ctx.Query("name")
		sharedToken := ctx.Query("shared_token")
		return ctx.SendString("name is : " + name + " " + sharedToken)
	})
	// Query parser
	app.Get("/query2", func(ctx *fiber.Ctx) error {
		person := model.Person{}
		ctx.QueryParser(&person)
		return ctx.JSON(person)
	})
	// Wildcard
	app.Get("/wildcard/*", func(ctx *fiber.Ctx) error {
		wildcard := ctx.Params("*")
		return ctx.SendString(wildcard)
	})
	// Static file
	app.Static("/", "../wwwroot")
	// New err
	app.Get("/error", func(ctx *fiber.Ctx) error {
		return fiber.NewError(fiber.StatusNotFound, "content not found")
	})

	// Group
	v1 := app.Group("/v1", func(ctx *fiber.Ctx) error {
		ctx.Set("Version", "V1")
		return ctx.Next()
	})
	v1.Get("/hello", func(ctx *fiber.Ctx) error {
		return ctx.SendString("Hello v1.")
	})
	v2 := app.Group("/v2")
	v2.Get("/hello", func(ctx *fiber.Ctx) error {
		return ctx.SendString("Hello v2.")
	})
	//Mount
	userApp := fiber.New()
	userApp.Get("/login", func(ctx *fiber.Ctx) error {
		return ctx.SendString("Login")
	})
	app.Mount("/user", userApp)
	// Sever
	app.Server().MaxConnsPerIP = 1 // max ip ได้แค่ 1 ip ถ้าอีก ip ยิงเข้ามาจะยิงไม่ได้
	app.Get("/server", func(ctx *fiber.Ctx) error {
		time.Sleep(time.Second * 15) //ตั้ง เวลารอโหลด ก่อนแสดงข้อมูล 15 วิ
		return ctx.SendString("server")
	})

	app.Get("/env", func(ctx *fiber.Ctx) error {
		return ctx.JSON(fiber.Map{
			"BaseURL":    ctx.BaseURL(),
			"Hostname":   ctx.Hostname(),
			"IP":         ctx.IP(),
			"IPs":        ctx.IPs(),
			"Protocal":   ctx.Protocol(),
			"Subdomains": ctx.Subdomains(),
		})
	})
	//Body
	app.Post("/body", func(ctx *fiber.Ctx) error {
		//fmt.Printf("IsJSON : %v\n", ctx.Is("json"))
		//fmt.Println(string(ctx.Body()))
		person := model.Person{}
		err := ctx.BodyParser(&person)
		if err != nil {
			return err
		}
		fmt.Println(person)
		return ctx.JSON(person)
	})
	//Body2
	app.Post("/body2", func(ctx *fiber.Ctx) error {
		//fmt.Printf("IsJSON : %v\n", ctx.Is("json"))
		//fmt.Println(string(ctx.Body()))
		data := map[string]interface{}{}
		err := ctx.BodyParser(&data)
		if err != nil {
			return err
		}
		fmt.Println(data)
		return ctx.JSON(data)
	})

	app.Listen(":8080")
}
