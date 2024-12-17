package horde

import (
	"fmt"
	"image/png"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/image/webp"
)

const (
	defaultSteps  = 30
	defaultWidth  = 640
	defaultHeight = 512
)

// ImageGenerate provides a simple way to generate a single image.
// It handles the entire process from request to download.
//
// Parameters:
// - prompt: The image description
// - steps: Number of inference steps (use 0 for default: 30)
// - width: Image width (use 0 for default: 640)
// - height: Image height (use 0 for default: 512)
//
// Returns:
// - []byte: The generated image data
// - error: Any error that occurred during the process
func (c *Client) ImageGenerate(prompt string, steps, width, height int, modelName string) ([]byte, error) {
	log.Printf("Starting image generation: prompt=%q, steps=%d, width=%d, height=%d",
		prompt, steps, width, height)

	// Apply defaults and log them
	if steps == 0 {
		steps = defaultSteps
		log.Printf("Using default steps: %d", steps)
	}
	if width == 0 {
		width = defaultWidth
		log.Printf("Using default width: %d", width)
	}
	if height == 0 {
		height = defaultHeight
		log.Printf("Using default height: %d", height)
	}
	if modelName == "" {
		modelName = defaultModel
		log.Printf("Using default model: %s", modelName)
	}

	// Create generation request
	req := GenerationRequest{
		Prompt: prompt,
		Params: Params{
			Steps:     steps,
			Width:     width,
			Height:    height,
			ModelName: modelName,
		},
	}

	// Request generation
	log.Printf("Submitting generation request...")
	resp, err := c.RequestGeneration(req)
	if err != nil {
		return nil, fmt.Errorf("requesting generation: %w", err)
	}
	log.Printf("Request accepted, got response: %v", resp)
	log.Printf("Request accepted, got ID: %s", resp.ID)

	// Wait for completion
	log.Printf("Waiting for generation to complete...")
	status, err := c.WaitForCompletion(resp.ID)
	if err != nil {
		return nil, fmt.Errorf("waiting for completion: %w", err)
	}

	log.Printf("Status: %v", status.Generation[0].Image)
	// Verify we have results
	/*if len(status.Generation) != 0 {
		return nil, fmt.Errorf("no results returned")
	}*/

	// Download the image
	log.Printf("Downloading generated image...")
	imageData, err := c.DownloadImage(status.Generation[0].Image)
	if err != nil {
		return nil, fmt.Errorf("downloading image: %w", err)
	}
	log.Printf("Successfully downloaded image: %d bytes", len(imageData))

	return imageData, nil
}

func Webp2PNG(input string) error {
	if input == "" {
		return fmt.Errorf("Error: No input file specified")
	}

	output := filepath.Base(input)
	output = output[0 : len(output)-len(filepath.Ext(output))] // Remove file extension

	f, err := os.Open(input)
	if err != nil {
		return fmt.Errorf("Error:", err)
	}
	defer f.Close()

	img, err := webp.Decode(f)
	if err != nil {
		return fmt.Errorf("Error:", err)
	}

	// Convert to PNG
	pngFile, err := os.Create(output + ".png")
	if err != nil {
		return fmt.Errorf("Error:", err)
	}
	defer pngFile.Close()

	err = png.Encode(pngFile, img)
	if err != nil {
		return fmt.Errorf("Error:", err)
	}

	fmt.Println("Conversion completed successfully")
	return nil
}
