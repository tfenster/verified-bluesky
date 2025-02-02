package shared

import (
	"fmt"

	"github.com/antchfx/htmlquery"
)

func HtmlXpathVerification(url string, xpathQuery string, bskyHandle string) (bool, error) {
	resp, err := SendGet(url, "")
	if err != nil {
		fmt.Println("Error fetching the URL: " + err.Error())
		return false, fmt.Errorf("Error fetching the HTML profile at " + url + ": " + err.Error())
	}
	defer resp.Body.Close()

	doc, err := htmlquery.Parse(resp.Body)
	if err != nil {
		fmt.Println("Error parsing HTML:", err)
		return false, fmt.Errorf("Error parsing the HTML profile at " + url + ": " + err.Error())
	}

	fmt.Println("XPath query: " + xpathQuery)
	nodes, err := htmlquery.QueryAll(doc, xpathQuery)
	if err != nil {
		fmt.Println("Error performing XPath query: %v", err)
		return false, fmt.Errorf("Could not find Bluesky URL https://bsky.app/profile/" + bskyHandle + " on the HTML profile at " + url + ": " + err.Error())
	}

	if len(nodes) == 0 {
		fmt.Println("Could not find Bluesky URL https://bsky.app/profile/" + bskyHandle + " on the HTML profile at " + url)
		return false, fmt.Errorf("Could not find Bluesky URL https://bsky.app/profile/" + bskyHandle + " on the HTML profile at " + url)
	}
	return true, nil
}
