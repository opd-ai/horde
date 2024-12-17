package horde

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	baseURL      = "https://stablehorde.net/api/v2"
	defaultWait  = 2 * time.Second
	defaultModel = "stable_diffusion_2.1"
)

// Client represents a Stable Horde API client
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new Stable Horde client
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GenerationRequest represents an image generation request
type GenerationRequest struct {
	Prompt string `json:"prompt"`
	Params Params `json:"params"`
}

type Params struct {
	Steps     int    `json:"steps"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	ModelName string `json:"model"`
}

// GenerationResponse represents the initial generation response
type GenerationResponse struct {
	ID     string  `json:"id"`
	Status string  `json:"status,omitempty"`
	Kudos  float64 `json:"kudos"`
}

// GenerationStatus represents the status check response
type GenerationStatus struct {
	Done       bool   `json:"done"`
	Failed     bool   `json:"failed"`
	Message    string `json:"message,omitempty"`
	WaitTime   int    `json:"wait_time"`      // Expected waiting time
	QueuePos   int    `json:"queue_position"` // Position in queue
	Processing int    `json:"processing"`     // Whether currently processing
	//Results    []Result    `json:"results"`
	Generation []Generations `json:"generations"`
}

type GenerationResult struct {
	Generation Generations `json:"generations"`
}

type Generations struct {
	Image string `json:"img"`
}

// Result represents a generated image result
type Result struct {
	ImageURL string `json:"img"`
}

// RequestGeneration initiates an image generation request
func (c *Client) RequestGeneration(req GenerationRequest) (*GenerationResponse, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	request, err := http.NewRequest("POST", baseURL+"/generate/async", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("apikey", c.apiKey)

	resp, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	// Change: Accept 202 status code
	if resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var genResp GenerationResponse
	if err := json.NewDecoder(resp.Body).Decode(&genResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &genResp, nil
}

// CheckStatus checks the status of a generation request
func (c *Client) CheckStatus(id string) (*GenerationStatus, error) {
	request, err := http.NewRequest("GET", baseURL+"/generate/check/"+id, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var status GenerationStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &status, nil
}

// CheckStatus checks the status of a generation request
func (c *Client) CheckRealStatus(id string) (*GenerationStatus, error) {
	request, err := http.NewRequest("GET", baseURL+"/generate/status/"+id, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unreadable response %s", err)
	}

	var status GenerationStatus
	if err := json.Unmarshal(body, &status); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &status, nil
}

// DownloadImage downloads a generated image
func (c *Client) DownloadImage(url string) ([]byte, error) {
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("downloading image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// WaitForCompletion waits for a generation request to complete
func (c *Client) WaitForCompletion(id string) (*GenerationStatus, error) {
	startTime := time.Now()
	attempts := 0

	for {
		attempts++
		status, err := c.CheckStatus(id)
		if err != nil {
			return nil, err
		}

		elapsed := time.Since(startTime)

		// Log detailed status
		log.Printf("Status check #%d [%s elapsed]: Queue Position: %d, Processing: %v, Wait Time: %ds",
			attempts,
			elapsed.Round(time.Second),
			status.QueuePos,
			status.Processing,
			status.WaitTime,
		)

		if status.Failed {
			return nil, fmt.Errorf("generation failed: %s", status.Message)
		}

		if status.Done {
			log.Printf("Generation completed after %s", elapsed.Round(time.Second))
			status, err := c.CheckRealStatus(id)
			if err != nil {
				return nil, err
			}
			return status, nil
		}

		// If processing, check more frequently
		if status.Processing != 0 {
			time.Sleep(2 * time.Second)
		} else {
			time.Sleep(5 * time.Second)
		}
	}
}
