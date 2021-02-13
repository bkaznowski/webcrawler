package crawler

import (
	"testing"
	"net/url"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCleanURL(t *testing.T) {
	var tests = []struct {
		name string
		inputURL string
		want string
	}{
		{"no changes required", "http://monzo.com","http://monzo.com"},
		{"removes query parameters", "http://monzo.com?param=123","http://monzo.com"},
		{"removes fragments", "http://monzo.com#fragment","http://monzo.com"},
		{"removes query parameters and fragments", "http://monzo.com?param=123#fragment","http://monzo.com"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			u, err := url.Parse(tc.inputURL)
			require.NoError(t, err, "invalid input URL", tc.inputURL)
			cleaned := cleanUrl(*u)
			assert.Equal(t, tc.want, cleaned)
		})
	}
	
}

func TestCrawler(t *testing.T) {
	t.Run("does not visit already visited links but does include links in output", func(t *testing.T) {
		mockedURLFinder := mockURLFinder{}

		c := newCrawler(4, &mockedURLFinder)
		parent := mustParseURL(t, "https://monzo.com")
		parentChildren := []*url.URL{
			mustParseURL(t, "https://monzo.com"),
			mustParseURL(t, "https://monzo.com"),
		}
		mockedURLFinder.On("find", parent).Return(parentChildren, nil).Once()

		results, err := c.Crawl("https://monzo.com")
		assert.NoError(t, err)

		expctedResults := []Result{
			{Parent: parent, Children: parentChildren},
		}
		assert.Equal(t, expctedResults, results)
	})

	t.Run("only visits sites with parameters once but include all found", func(t *testing.T) {
		mockedURLFinder := mockURLFinder{}

		c := newCrawler(4, &mockedURLFinder)
		parent := mustParseURL(t, "https://monzo.com")
		firstParamSite := mustParseURL(t, "https://monzo.com/params?param=1")
		secondParamSite := mustParseURL(t, "https://monzo.com/params?param=2")
		parentChildren := []*url.URL{firstParamSite, secondParamSite}
		mockedURLFinder.On("find", parent).Return(parentChildren, nil).Once()
		mockedURLFinder.On("find", firstParamSite).Return([]*url.URL{}, nil).Once()

		results, err := c.Crawl("https://monzo.com")
		assert.NoError(t, err)

		expctedResults := []Result{
			{Parent: parent, Children: parentChildren},
			{Parent: firstParamSite, Children: []*url.URL{}},
		}
		assert.Equal(t, expctedResults, results)
	})

	t.Run("removes fragments and only visits once", func(t *testing.T) {
		mockedURLFinder := mockURLFinder{}

		c := newCrawler(4, &mockedURLFinder)
		parent := mustParseURL(t, "https://monzo.com")
		firstParamSite := mustParseURL(t, "https://monzo.com/params#frag1")
		secondParamSite := mustParseURL(t, "https://monzo.com/params#frag2")
		parentChildren := []*url.URL{firstParamSite, secondParamSite}
		mockedURLFinder.On("find", parent).Return(parentChildren, nil).Once()
		mockedURLFinder.On("find", firstParamSite).Return([]*url.URL{}, nil).Once()

		results, err := c.Crawl("https://monzo.com")
		assert.NoError(t, err)

		expctedResults := []Result{
			{Parent: parent, Children: parentChildren},
			{Parent: firstParamSite, Children: []*url.URL{}},
		}
		assert.Equal(t, expctedResults, results)
	})

	t.Run("handles links when only path provided", func(t *testing.T) {
		mockedURLFinder := mockURLFinder{}

		c := newCrawler(4, &mockedURLFinder)
		parent := mustParseURL(t, "https://monzo.com")
		firstPathSite := mustParseURL(t, "/test1")
		secondPathSite := mustParseURL(t, "/test2/test")
		firstSite := mustParseURL(t, "https://monzo.com/test1")
		secondSite := mustParseURL(t, "https://monzo.com/test2/test")
		parentChildren := []*url.URL{firstPathSite, secondPathSite}
		mockedURLFinder.On("find", parent).Return(parentChildren, nil).Once()
		mockedURLFinder.On("find", firstSite).Return([]*url.URL{}, nil).Once()
		mockedURLFinder.On("find", secondSite).Return([]*url.URL{}, nil).Once()

		results, err := c.Crawl("https://monzo.com")
		assert.NoError(t, err)

		expctedResults := []Result{
			{Parent: parent, Children: parentChildren},
			{Parent: firstSite, Children: []*url.URL{}},
			{Parent: secondSite, Children: []*url.URL{}},
		}
		assert.ElementsMatch(t, expctedResults, results)
	})

}

func mustParseURL(t *testing.T, u string) *url.URL {
	parsed, err := url.Parse(u)
	require.NoError(t, err)
	return parsed
}