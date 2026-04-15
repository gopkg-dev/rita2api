package upstream

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL     string
	origin      string
	language    string
	modelTypeID int
	modelID     int
	httpClient  *http.Client
}

type GenerateRequest struct {
	Prompt     string
	Ratio      string
	Resolution string
	ImageNum   int
}

type GenerateResponse struct {
	ParentMessageID string
	SSEPath         string
}

type generatePayload struct {
	Language    string `json:"language"`
	ModelTypeID int    `json:"model_type_id"`
	Generate    struct {
		ModelID        int    `json:"model_id"`
		ImageNum       int    `json:"image_num"`
		Ratio          string `json:"ratio"`
		ImageLike      int    `json:"image_like"`
		Resolution     string `json:"resolution"`
		Prompt         string `json:"prompt"`
		GenerationMode string `json:"generation_mode"`
	} `json:"generate"`
}

type generateAPIResponse struct {
	Data struct {
		SSEPath         string `json:"sse_path"`
		ParentMessageID string `json:"parent_message_id"`
	} `json:"data"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func NewClient(baseURL, origin string, modelTypeID, modelID int) *Client {
	return &Client{
		baseURL:     strings.TrimRight(baseURL, "/"),
		origin:      strings.TrimRight(origin, "/"),
		language:    "zh",
		modelTypeID: modelTypeID,
		modelID:     modelID,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (c *Client) Generate(ctx context.Context, visitorHeader string, req GenerateRequest) (GenerateResponse, error) {
	var payload generatePayload
	payload.Language = c.language
	payload.ModelTypeID = c.modelTypeID
	payload.Generate.ModelID = c.modelID
	payload.Generate.ImageNum = req.ImageNum
	payload.Generate.Ratio = req.Ratio
	payload.Generate.ImageLike = 4
	payload.Generate.Resolution = req.Resolution
	payload.Generate.Prompt = req.Prompt
	payload.Generate.GenerationMode = "text-to-image"

	body, err := json.Marshal(payload)
	if err != nil {
		return GenerateResponse{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/aiart/generate", bytes.NewReader(body))
	if err != nil {
		return GenerateResponse{}, err
	}
	c.applyHeaders(httpReq, visitorHeader)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return GenerateResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return GenerateResponse{}, fmt.Errorf("rita generate: %s", strings.TrimSpace(string(raw)))
	}

	var decoded generateAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return GenerateResponse{}, err
	}

	if decoded.Code != 0 {
		return GenerateResponse{}, fmt.Errorf("rita generate failed: %s", decoded.Message)
	}

	return GenerateResponse{
		ParentMessageID: decoded.Data.ParentMessageID,
		SSEPath:         decoded.Data.SSEPath,
	}, nil
}

func (c *Client) Stream(ctx context.Context, visitorHeader, parentMessageID string) ([]StreamEvent, error) {
	url := c.baseURL + "/ai/task/record/push?parent_message_id=" + parentMessageID
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	c.applyHeaders(httpReq, visitorHeader)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("rita stream: %s", strings.TrimSpace(string(raw)))
	}

	return ParseStreamMessages(resp.Body)
}

func (c *Client) applyHeaders(req *http.Request, visitorHeader string) {
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Origin", c.origin)
	req.Header.Set("Referer", c.origin+"/")
	req.Header.Set("User-Agent", "rati-ai-studio/1.0")
	req.Header.Set("VisitorId", visitorHeader)
}
