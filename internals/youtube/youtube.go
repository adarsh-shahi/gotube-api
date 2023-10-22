package youtube

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Channel struct {
	Title           string `json:"name"`
	CustomUrl       string `json:"username"`
	Description     string `json:"description"`
	ProfileImageUrl string `json:"profileImageUrl"`
}
type Response struct {
	Items []struct {
		Snippet struct {
			CustomURL   string `json:"customUrl"`
			Description string `json:"description"`
			Title       string `json:"title"`
			Thumbnails  struct {
				Medium struct {
					Url string `json:"url"`
				}
			}
		} `json:"snippet"`
	} `json:"items"`
}

const (
	channelUrl = "https://www.googleapis.com/youtube/v3/channels"
)

func GetChannelInfo(accessToken string) (*Channel, error) {

	client := &http.Client{}
	req, _ := http.NewRequest("GET", channelUrl+"?mine=true&part=snippet", nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	response := Response{}
	json.NewDecoder(resp.Body).Decode(&response)
	ch := &Channel{}
	ch.Title = response.Items[0].Snippet.Title
	ch.Description = response.Items[0].Snippet.Description
	ch.CustomUrl = response.Items[0].Snippet.CustomURL
	ch.ProfileImageUrl = response.Items[0].Snippet.Thumbnails.Medium.Url
	return ch, nil
}
