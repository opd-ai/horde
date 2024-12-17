package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/opd-ai/horde"
)

var (
	prompt    = flag.String("prompt", "", "prompt")
	height    = flag.Int("height", 512, "image height")
	width     = flag.Int("width", 640, "image width")
	modelName = flag.String("modelname", "stable_diffusion_2.1", "model to useu")
	steps     = flag.Int("steps", 30, "steps to generate")
	png       = flag.Bool("png", true, "conveert output to png")
	output    = flag.String("output", "image.webp", "output filename")
)

func main() {
	apiKey := os.Getenv("HORDE_API_KEY")
	if apiKey == "" {
		log.Fatal("Set a stable horde API key")
	}
	if len(*prompt) < 5 {
		flag.Usage()
		fmt.Printf("> Error: %s\n", "You must enter a prompt")
		os.Exit(1)
	}
	client := horde.NewClient(apiKey)
	if imageData, err := client.ImageGenerate(*prompt, *steps, *width, *height, *modelName); err != nil {
		log.Fatal(err)
	} else {
		if err := os.WriteFile(*output, imageData, 0o644); err != nil {
			log.Fatal(err)
		} else {
			if *png {
				if err := horde.Webp2PNG(*output); err != nil {
					log.Fatal(err)
				}
			}
		}
	}
}
