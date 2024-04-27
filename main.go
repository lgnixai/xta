package main

import (
	"context"
	"encoding/json"
	`fmt`
	`io`
	`log`
	`log/slog`
	"net/http"
	`time`

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/prompts"
)

func main() {
	r := gin.Default()
	//r.Use(cors.New(cors.Config{
	//	AllowOrigins:     []string{"*"},
	//	AllowMethods:     []string{"PUT", "PATCH"},
	//	AllowHeaders:     []string{"Origin"},
	//	ExposeHeaders:    []string{"Content-Length"},
	//	AllowCredentials: true,
	//	AllowOriginFunc: func(origin string) bool {
	//		return origin == "*"
	//	},
	//	MaxAge: 12 * time.Hour,
	//}))
	r.Use(cors.Default())
	v1 := r.Group("/api/v1")

	v1.POST("/translator", translator)
	v1.POST("/fanyi", fanyi)

	r.Run(":8888")
}

func study(ctx context.Context) {
	llm, err := ollama.New(ollama.WithModel("mistral"))
	if err != nil {
		log.Fatal(err)
	}
	//ctx := context.Background()

	content := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, "You are a company branding design wizard."),
		llms.TextParts(llms.ChatMessageTypeHuman, "What would be a good company name a company that makes colorful socks?"),
	}
	completion, err := llm.GenerateContent(ctx, content, llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
		fmt.Print(string(chunk))
		return nil
	}))
	if err != nil {
		log.Fatal(err)
	}
	_ = completion
}
func ok(c *gin.Context) {
	ticker := time.NewTicker(1 * time.Second)
	i := 0
	// 子协程
	go func() {
		for {
			//<-ticker.C
			i++
			fmt.Println(<-ticker.C)
			if i == 5 {
				//停止
				ticker.Stop()
			}
		}
	}()
}

type User struct {
	Context  string `json:"context"`
	Done     bool   `json:"done"`
	Response string `json:"response"`
}

func fanyi(c *gin.Context) {
	llm, err := ollama.New(ollama.WithModel("mistral"))
	if err != nil {
		log.Fatal(err)
	}
	//ctx := context.Background()

	content := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, "你是一位精通简体中文的专业翻译 ,现在请翻译以下内容为简体中文：\n"),
		llms.TextParts(llms.ChatMessageTypeHuman, "What would be a good company name a company that makes colorful socks?"),
	}
	ch := make(chan any)
	defer close(ch)
	ctx := context.Background()

	//
	llm.GenerateContent(ctx, content, llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
		go func() {
			//resp := (chunk)
			//fmt.Print(string(chunk) + "\n")

			fmt.Print(string(chunk) + "\n")
			user := User{Done: false, Context: "10", Response: string(chunk)}

			ch <- user

		}()
		return nil
	}))

	user := User{Done: true, Context: "10", Response: "string(chunk)"}
	ch <- user
	streamResponse(c, ch)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//c.JSON(http.StatusOK, completion.Choices[0].Content)
	//return
	//_ = completion

	//ch := make(chan any)
	//defer close(ch)
	//
	//ticker := time.NewTicker(1 * time.Second)
	//i := 0
	//go func() {
	//	//defer close(ch)
	//	for {
	//		<-ticker.C
	//		i = i + 1
	//		if i == 6 {
	//			user := User{Done: true, Context: "10", Response: fmt.Sprintf("%d", i)}
	//
	//			ch <- user
	//
	//			break
	//		} else {
	//			user := User{Done: false, Context: "10", Response: fmt.Sprintf("%d", i)}
	//
	//			ch <- user
	//		}
	//	}
	//}()

	//
	//fmt.Println(ch)
	//for resp := range ch {
	//
	//	fmt.Println(resp)
	//	switch r := resp.(type) {
	//	case api.GenerateResponse:
	//		sb.WriteString(r.Response)
	//		final = r
	//	case gin.H:
	//		if errorMsg, ok := r["error"].(string); ok {
	//			c.JSON(http.StatusInternalServerError, gin.H{"error": errorMsg})
	//			return
	//		} else {
	//			c.JSON(http.StatusInternalServerError, gin.H{"error": "unexpected error format in response"})
	//			return
	//		}
	//	default:
	//		c.JSON(http.StatusInternalServerError, gin.H{"error": "unexpected error"})
	//		return
	//	}
	//}
	//
	//final.Response = sb.String()
	//c.JSON(http.StatusOK, final)
	//return
	////if req.Stream != nil && !*req.Stream {
	////	waitForStream(c, ch)
	////	return
	////}
	//
	//streamResponse(c, ch)
}

func waitForStream(c *gin.Context, ch chan interface{}) {
	c.Header("Content-Type", "application/json")
	for resp := range ch {
		switch r := resp.(type) {

		case gin.H:
			if errorMsg, ok := r["error"].(string); ok {
				c.JSON(http.StatusInternalServerError, gin.H{"error": errorMsg})
				return
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "unexpected error format in progress response"})
				return
			}
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "unexpected progress response"})
			return
		}
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "unexpected end of progress response"})
}

func streamResponse(c *gin.Context, ch chan any) {
	c.Header("Content-Type", "application/x-ndjson")
	c.Stream(func(w io.Writer) bool {
		val, ok := <-ch
		fmt.Println("val", val.(User).Done, ok)
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

func fanyi2(c *gin.Context) {
	var requestData struct {
		Text string `json:"text"`
	}

	if err := c.BindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error:": "绑定json3333参数错误"})
		return
	}
	text := requestData.Text
	var ollamaServer = "http://localhost:11434"
	llm, err := ollama.New(
		ollama.WithModel("llama2-chinese:latest"),
		ollama.WithServerURL(ollamaServer))
	fmt.Println(err)
	fmt.Println(text)
	// Translate 将文本翻译为中文
	//func Translate(llm llms.Model, text string) (string, error) {
	response, err := llms.GenerateFromSinglePrompt(
		context.TODO(),
		llm,
		"To translate the following sentence into Chinese, you only need to reply to the translated content, and do not need to reply to any other content. The output format is json, and there are two fields: conten: the translated Chinese content, length: the length of the content, and the English content to be translated: \n"+text,
		llms.WithTemperature(0.8))

	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error:": err})
	}
	fmt.Println(response)

	c.JSON(http.StatusOK, gin.H{"response": response})
	//if err != nil {
	//	return "", err
	//}
	//return completion, nil
	//}
}

func translator(c *gin.Context) {
	var requestData struct {
		OutputLang string `json:"outputlang"`
		Text       string `json:"text"`
	}

	if err := c.BindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error:": "绑定json参数错误"})
		return
	}

	// 创建prompt
	prompt := prompts.NewChatPromptTemplate([]prompts.MessageFormatter{
		// 模型处理规则描述
		prompts.NewSystemMessagePromptTemplate("你是一个销售人员态度分析引擎,通过对销售人员的态度进行评分,如果销售人员在销售过程中对顾客有嘲讽等恶意情绪请结合语气给出评分,满分100分,你只需要输出评分和简单评价", nil),
		// 输入
		prompts.NewHumanMessagePromptTemplate(`{{.text}}:{{.outputlang}}`, []string{"text", "outputlang"}),
	})

	// 填充prompt
	value := map[string]any{
		"outputlang": requestData.OutputLang,
		"text":       requestData.Text,
	}

	messages, err := prompt.FormatMessages(value)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error:": "消息处理错误!"})
		return
	}

	//链接ollama
	llm, err := ollama.New(ollama.WithModel("qwen:7b"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error:": err})
	}

	content := []llms.MessageContent{
		llms.TextParts(messages[0].GetType(), messages[0].GetContent()),
		llms.TextParts(messages[1].GetType(), messages[1].GetContent()),
	}

	response, err := llm.GenerateContent(context.Background(), content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error:": err})
	}

	c.JSON(http.StatusOK, gin.H{"response": response.Choices[0].Content})
}
