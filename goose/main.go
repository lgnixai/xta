package main

import (
	"context"
	`encoding/json`
	"fmt"
	`io`
	`log/slog`
	`time`

	`github.com/gin-contrib/cors`
	`github.com/gin-gonic/gin`
	"github.com/henomis/lingoose/llm/ollama"
	"github.com/henomis/lingoose/thread"
	"github.com/ollama/ollama/api"
)

type User struct {
	Context  string `json:"context"`
	Done     bool   `json:"done"`
	Response string `json:"response"`
}

var mode string = gin.DebugMode

func init() {
	switch mode {
	case gin.DebugMode:
	case gin.ReleaseMode:
	case gin.TestMode:
	default:
		mode = gin.DebugMode
	}

	gin.SetMode(mode)
}
func main() {
	r := gin.Default()

	r.Use(cors.Default())
	v1 := r.Group("/api/v1")

	v1.POST("/fanyi", goose)

	r.Run(":8888")
}

func fanyi(c *gin.Context) {
	ch := make(chan any)
	defer close(ch)

	ticker := time.NewTicker(1 * time.Second)
	i := 0
	go func() {
		//defer close(ch)
		for {
			<-ticker.C
			i = i + 1
			if i == 10 {
				user := User{Done: true, Context: "10", Response: fmt.Sprintf("%d", i)}

				ch <- user

				break
			} else {
				user := User{Done: false, Context: "10", Response: fmt.Sprintf("%d", i)}

				ch <- user
			}
		}
	}()
	streamResponse(c, ch)
}
func goose(c *gin.Context) {

	t1 := thread.New()
	t1.AddMessage(thread.NewUserMessage().AddContent(
		thread.NewTextContent("河南的省会是哪？"),
	))
	ch := make(chan any)

	go func() {
		defer close(ch)
		err := ollama.New().WithEndpoint("http://localhost:11434/api").WithModel("llama2-chinese:latest").
			WithStream(func(s string) {
				resp := api.GenerateResponse{
					Model:     "req.Model",
					CreatedAt: time.Now().UTC(),
					Response:  s,
					Done:      false,
				}
				//fmt.Print(s)
				//fmt.Print(s)
				//user := User{Done: false, Context: "10", Response: string(s)}

				ch <- resp

			}).Generate(context.Background(), t1)
		if err != nil {
			panic(err)
		}
		//fmt.Print("over")
		resp := api.GenerateResponse{
			Model:     "req.Model",
			CreatedAt: time.Now().UTC(),
			Response:  "",
			Done:      true,
		}

		ch <- resp
	}()
	fmt.Println(t1)

	//var final api.GenerateResponse
	//var sb strings.Builder
	//for resp := range ch {
	//
	//	switch r := resp.(type) {
	//	case api.GenerateResponse:
	//		fmt.Println(r.Response, "32432")
	//		sb.WriteString(r.Response)
	//		final = r
	//
	//	default:
	//		c.JSON(http.StatusInternalServerError, gin.H{"error": "unexpected error"})
	//		return
	//	}
	//}
	//
	//final.Response = sb.String()
	//c.JSON(http.StatusOK, final)
	//
	//return

	streamResponse(c, ch)
	//go func() {
	//	user := User{Done: true, Context: "10", Response: "string(chunk)"}
	//	ch <- user
	//}()

	//fmt.Println(t.String())

}
func streamResponse(c *gin.Context, ch chan any) {
	c.Header("Content-Type", "application/x-ndjson")
	c.Stream(func(w io.Writer) bool {
		val, ok := <-ch
		fmt.Println("val,ok", val, ok)
		if !ok {
			return false
		}

		bts, err := json.Marshal(val)
		fmt.Println(string(bts))
		if err != nil {
			fmt.Println(err)
			slog.Info(fmt.Sprintf("streamResponse: json.Marshal failed with %s", err))
			return false
		}

		// Delineate chunks with new-line delimiter
		bts = append(bts, '\n')

		if _, err := w.Write(bts); err != nil {
			slog.Info(fmt.Sprintf("streamResponse: w.Write failed with %s", err))
			return false
		}

		return true
	})
}
