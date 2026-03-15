package catalog

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// BOMItem matches the response from catalog-service GET /api/v1/catalog/menus/{id}/bom
type BOMItem struct {
	IngredientID uint    `json:"ingredient_id"`
	Quantity     float64 `json:"quantity"`
}

type bomResponse struct {
	BOMItems []BOMItem `json:"bom_items"`
}

// Client is a lightweight HTTP client for the catalog-service.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// GetBOM fetches the bill of materials for a menu item from catalog-service.
func (c *Client) GetBOM(menuItemID uint) ([]BOMItem, error) {
	url := fmt.Sprintf("%s/api/v1/catalog/menus/%d/bom", c.baseURL, menuItemID)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("catalog client: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("catalog client: unexpected status %d for menu %d", resp.StatusCode, menuItemID)
	}

	var body bomResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("catalog client: decode: %w", err)
	}
	return body.BOMItems, nil
}
