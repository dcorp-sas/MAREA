package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

const maxComfyResponseBytes = 32 * 1024 * 1024

// ComfyImages sends image-generation work to a ComfyUI instance on the
// private Marea network. It does not call a hosted model provider.
type ComfyImages struct {
	baseURL    *url.URL
	checkpoint string
	http       *http.Client
}

var _ ImageProvider = (*ComfyImages)(nil)

func NewComfyImages(baseURL, checkpoint string) (*ComfyImages, error) {
	base, err := url.Parse(strings.TrimRight(strings.TrimSpace(baseURL), "/"))
	if err != nil || base == nil || (base.Scheme != "http" && base.Scheme != "https") || base.Host == "" || base.User != nil || base.RawQuery != "" || base.Fragment != "" {
		return nil, errors.New("invalid ComfyUI base URL")
	}
	checkpoint = strings.TrimSpace(checkpoint)
	if checkpoint == "" || strings.ContainsAny(checkpoint, "/\\") || checkpoint == "." || checkpoint == ".." {
		return nil, errors.New("invalid ComfyUI checkpoint name")
	}
	return &ComfyImages{
		baseURL:    base,
		checkpoint: checkpoint,
		http:       &http.Client{Timeout: 90 * time.Second, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }},
	}, nil
}

func (c *ComfyImages) SubmitImage(ctx context.Context, input ImageRequest) (string, error) {
	workflow, err := c.workflow(ctx, input)
	if err != nil {
		return "", err
	}
	body, err := json.Marshal(map[string]any{"prompt": workflow, "client_id": "marea"})
	if err != nil {
		return "", errors.New("ComfyUI workflow could not be encoded")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint("prompt"), bytes.NewReader(body))
	if err != nil {
		return "", errors.New("ComfyUI request could not be created")
	}
	request.Header.Set("Content-Type", "application/json")
	request.GetBody = nil
	response, err := c.http.Do(request)
	if err != nil {
		return "", errors.New("ComfyUI did not return a generation receipt")
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", errors.New("ComfyUI rejected the image request")
	}
	var receipt struct {
		PromptID string `json:"prompt_id"`
	}
	if json.NewDecoder(io.LimitReader(response.Body, 64*1024)).Decode(&receipt) != nil || !validImageRequestID(receipt.PromptID) {
		return "", errors.New("ComfyUI did not return a valid generation receipt")
	}
	return receipt.PromptID, nil
}

func (c *ComfyImages) workflow(ctx context.Context, input ImageRequest) (map[string]any, error) {
	width, height := comfyDimensions(input.Size)
	workflow := map[string]any{
		"1": node("CheckpointLoaderSimple", map[string]any{"ckpt_name": c.checkpoint}),
		"2": node("CLIPTextEncode", map[string]any{"text": input.Prompt, "clip": []any{"1", 1}}),
		"3": node("CLIPTextEncode", map[string]any{"text": "low quality, blurry, distorted text, watermark", "clip": []any{"1", 1}}),
		"7": node("SaveImage", map[string]any{"filename_prefix": "marea", "images": []any{"6", 0}}),
	}
	if len(input.Source) > 0 {
		source, err := c.upload(ctx, "source.png", input.Source)
		if err != nil {
			return nil, err
		}
		workflow["8"] = node("LoadImage", map[string]any{"image": source})
		workflow["9"] = node("VAEEncode", map[string]any{"pixels": []any{"8", 0}, "vae": []any{"1", 2}})
		latent := []any{"9", 0}
		denoise := 0.45
		if len(input.Mask) > 0 {
			mask, err := c.upload(ctx, "selection.png", input.Mask)
			if err != nil {
				return nil, err
			}
			workflow["10"] = node("LoadImage", map[string]any{"image": mask})
			workflow["11"] = node("InvertMask", map[string]any{"mask": []any{"10", 1}})
			workflow["12"] = node("SetLatentNoiseMask", map[string]any{"samples": latent, "mask": []any{"11", 0}})
			latent = []any{"12", 0}
			denoise = 0.85
		}
		workflow["5"] = node("KSampler", map[string]any{"seed": time.Now().UnixNano() & 0x7fffffffffffffff, "steps": 24, "cfg": 6.5, "sampler_name": "dpmpp_2m", "scheduler": "karras", "denoise": denoise, "model": []any{"1", 0}, "positive": []any{"2", 0}, "negative": []any{"3", 0}, "latent_image": latent})
	} else {
		workflow["4"] = node("EmptyLatentImage", map[string]any{"width": width, "height": height, "batch_size": 1})
		workflow["5"] = node("KSampler", map[string]any{"seed": time.Now().UnixNano() & 0x7fffffffffffffff, "steps": 24, "cfg": 6.5, "sampler_name": "dpmpp_2m", "scheduler": "karras", "denoise": 1.0, "model": []any{"1", 0}, "positive": []any{"2", 0}, "negative": []any{"3", 0}, "latent_image": []any{"4", 0}})
	}
	workflow["6"] = node("VAEDecode", map[string]any{"samples": []any{"5", 0}, "vae": []any{"1", 2}})
	return workflow, nil
}

func node(classType string, inputs map[string]any) map[string]any {
	return map[string]any{"class_type": classType, "inputs": inputs}
}

func comfyDimensions(size string) (int, int) {
	switch size {
	case "portrait_4_3":
		return 896, 1152
	case "portrait_16_9":
		return 768, 1344
	case "landscape_4_3":
		return 1152, 896
	case "landscape_16_9":
		return 1344, 768
	default:
		return 1024, 1024
	}
}

func (c *ComfyImages) upload(ctx context.Context, filename string, data []byte) (string, error) {
	if len(data) == 0 || len(data) > maxImageBytes {
		return "", errors.New("source image is empty or too large for ComfyUI")
	}
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("image", filename)
	if err != nil {
		return "", errors.New("ComfyUI upload could not be prepared")
	}
	if _, err = part.Write(data); err != nil {
		return "", errors.New("ComfyUI upload could not be prepared")
	}
	_ = writer.WriteField("type", "input")
	_ = writer.WriteField("overwrite", "true")
	if err = writer.Close(); err != nil {
		return "", errors.New("ComfyUI upload could not be prepared")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint("upload/image"), &body)
	if err != nil {
		return "", errors.New("ComfyUI upload request could not be created")
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())
	response, err := c.http.Do(request)
	if err != nil {
		return "", errors.New("ComfyUI source image upload failed")
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", errors.New("ComfyUI rejected the source image")
	}
	var uploaded struct {
		Name      string `json:"name"`
		Subfolder string `json:"subfolder"`
		Type      string `json:"type"`
	}
	if json.NewDecoder(io.LimitReader(response.Body, 64*1024)).Decode(&uploaded) != nil || uploaded.Name == "" || strings.ContainsAny(uploaded.Name, "/\\") || strings.ContainsAny(uploaded.Subfolder, "\\") {
		return "", errors.New("ComfyUI returned an invalid uploaded image reference")
	}
	return path.Join(uploaded.Subfolder, uploaded.Name), nil
}

func (c *ComfyImages) ImageStatus(ctx context.Context, _ string, id string) (ImageStatus, error) {
	if !validImageRequestID(id) {
		return ImageStatus{}, errors.New("invalid ComfyUI receipt")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint("history/"+url.PathEscape(id)), nil)
	if err != nil {
		return ImageStatus{}, errors.New("ComfyUI status request could not be created")
	}
	response, err := c.http.Do(request)
	if err != nil {
		return ImageStatus{}, errors.New("ComfyUI status is unavailable")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return ImageStatus{}, fmt.Errorf("ComfyUI status is unavailable (%d)", response.StatusCode)
	}
	var history map[string]struct {
		Status struct {
			StatusStr string `json:"status_str"`
			Completed bool   `json:"completed"`
		} `json:"status"`
		Outputs map[string]any `json:"outputs"`
	}
	if json.NewDecoder(io.LimitReader(response.Body, 2*1024*1024)).Decode(&history) != nil {
		return ImageStatus{}, errors.New("ComfyUI returned an invalid status")
	}
	entry, ok := history[id]
	if !ok {
		return ImageStatus{State: "generating"}, nil
	}
	if strings.EqualFold(entry.Status.StatusStr, "error") {
		return ImageStatus{State: "failed", Failed: true}, nil
	}
	if entry.Status.Completed || strings.EqualFold(entry.Status.StatusStr, "success") {
		return ImageStatus{State: "completed"}, nil
	}
	return ImageStatus{State: "generating"}, nil
}

func (c *ComfyImages) ImageResult(ctx context.Context, _ string, id string) ([]byte, error) {
	if !validImageRequestID(id) {
		return nil, errors.New("invalid ComfyUI receipt")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint("history/"+url.PathEscape(id)), nil)
	if err != nil {
		return nil, errors.New("ComfyUI result request could not be created")
	}
	response, err := c.http.Do(request)
	if err != nil {
		return nil, errors.New("ComfyUI result is unavailable")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, errors.New("ComfyUI result is unavailable")
	}
	var history map[string]struct {
		Outputs map[string]struct {
			Images []struct {
				Filename  string `json:"filename"`
				Subfolder string `json:"subfolder"`
				Type      string `json:"type"`
			} `json:"images"`
		} `json:"outputs"`
	}
	if json.NewDecoder(io.LimitReader(response.Body, 2*1024*1024)).Decode(&history) != nil {
		return nil, errors.New("ComfyUI returned an invalid result")
	}
	entry, ok := history[id]
	if !ok {
		return nil, errors.New("ComfyUI result is not ready")
	}
	for _, output := range entry.Outputs {
		for _, image := range output.Images {
			if image.Filename == "" || strings.ContainsAny(image.Filename, "/\\") || strings.ContainsAny(image.Subfolder, "\\") {
				continue
			}
			query := url.Values{"filename": {image.Filename}, "subfolder": {image.Subfolder}, "type": {image.Type}}
			resultRequest, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint("view")+"?"+query.Encode(), nil)
			if err != nil {
				return nil, errors.New("ComfyUI image download could not be created")
			}
			result, err := c.http.Do(resultRequest)
			if err != nil {
				return nil, errors.New("ComfyUI image download failed")
			}
			defer result.Body.Close()
			if result.StatusCode != http.StatusOK {
				return nil, errors.New("ComfyUI image download failed")
			}
			data, err := io.ReadAll(io.LimitReader(result.Body, maxComfyResponseBytes+1))
			if err != nil || len(data) == 0 || len(data) > maxComfyResponseBytes {
				return nil, errors.New("ComfyUI image exceeds the download limit")
			}
			return data, nil
		}
	}
	return nil, errors.New("ComfyUI completed without an image")
}

func (c *ComfyImages) ImageCost(context.Context, string, string) (*int64, error) {
	zero := int64(0)
	return &zero, nil
}

func (c *ComfyImages) endpoint(relative string) string {
	result := *c.baseURL
	result.Path = path.Join(strings.TrimSuffix(result.Path, "/"), relative)
	result.RawPath = ""
	return result.String()
}
