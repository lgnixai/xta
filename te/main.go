package main

import (
	`context`
	`fmt`
	`log`

	`github.com/tmc/langchaingo/llms`
	`github.com/tmc/langchaingo/llms/ollama`
)

func main() {
	llm, err := ollama.New(ollama.WithModel("mistral"))
	if err != nil {
		log.Fatal(err)
	}
	//ctx := context.Background()

	content := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, "你是知识达人：\n"),
		llms.TextParts(llms.ChatMessageTypeHuman, "河南省会是哪里?"),
	}
	ch := make(chan any)
	defer close(ch)
	ctx := context.Background()

	//
	completion, err := llm.GenerateContent(ctx, content, llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {

		//resp := (chunk)
		fmt.Print(string(chunk))
		//go func() {
		//	fmt.Print(string(chunk))
		//	//user := User{Done: false, Context: "10", Response: string(chunk)}
		//
		//	//ch <- user
		//}()
		return nil
	}))
	fmt.Println("=====")
	fmt.Println(completion.Choices[0].Content)
	_ = completion

}
