package main

import (
	"context"
	"fmt"
	"os"

	"github.com/henomis/lingoose/assistant"
	ollamaembedder `github.com/henomis/lingoose/embedder/ollama`
	"github.com/henomis/lingoose/index"
	"github.com/henomis/lingoose/index/vectordb/jsondb"
	"github.com/henomis/lingoose/llm/openai"
	"github.com/henomis/lingoose/rag"
	"github.com/henomis/lingoose/thread"

	goopenai "github.com/sashabaranov/go-openai"
)

// download https://raw.githubusercontent.com/hwchase17/chat-your-data/master/state_of_the_union.txt

func main() {
	customConfig := goopenai.DefaultConfig("")
	customConfig.BaseURL = "http://localhost:11434/v1"
	customClient := goopenai.NewClientWithConfig(customConfig)
	llm := openai.New().WithClient(customClient).WithModel("gemma:latest")
	r := rag.NewSubDocument(
		index.New(
			jsondb.New().WithPersist("db.json"),
			ollamaembedder.New().WithModel("nomic-embed-text"),

		),
		llm,
	).WithTopK(3)

	_, err := os.Stat("db.json")
	if os.IsNotExist(err) {
		err = r.AddSources(context.Background(), "./1.txt")
		if err != nil {
			panic(err)
		}
	}

	a := assistant.New(
		llm.WithTemperature(0),
	).WithRAG(r).WithThread(
		thread.New().AddMessages(
			thread.NewUserMessage().AddContent(
				thread.NewTextContent("河南怎么样?"),
			),
		),
	)

	err = a.Run(context.Background())
	if err != nil {
		panic(err)
	}

	fmt.Println("----")
	fmt.Println(a.Thread())
	fmt.Println("----")
}
