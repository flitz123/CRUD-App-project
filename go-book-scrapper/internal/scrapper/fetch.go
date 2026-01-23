func fetchHTML(url string) (*http.Response, error) {
	client := &http.Client{
		Timrout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOk {
		return nil, fmt.Errorf("failed to fetch page: %s", resp.Status)
	}

	return resp, nil
}