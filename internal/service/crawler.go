package service

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/Icarus-xD/HabrParser/internal/model"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

const (
	BASE_URL = "https://habr.com"
	ARTICLE_URL_LEN = 36
	CRAWL_ITERATION_TIME = time.Minute * 10
	TIME_HUB_CRAWLS = time.Second
	TIME_BETWEEN_REQUESTS = time.Millisecond * 20
)

type LinkRepo interface {
	GetAll() ([]model.Link, error)
}

type HubRepo interface {
	Create(hubInfo model.Hub) (model.Hub, error)
}

type ArticleRepo interface {
	Create(articleInfo model.Article) error
}

type Crawler struct {
	httpClient *http.Client
	linkRepo LinkRepo
	hubRepo HubRepo
	articleRepo ArticleRepo
}

type RequestResult struct {
	URL string
	Doc string
	IsErr bool
}

type ArticleInfoResult struct {
	Info model.Article
	IsErr bool
}

func NewCrawler(link LinkRepo, hub HubRepo, article ArticleRepo) *Crawler {
	return &Crawler{
		httpClient: &http.Client{
			Transport: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
	},
		linkRepo: link,
		hubRepo: hub,
		articleRepo: article,
	}
}

func (s *Crawler) RunCrawling() error {
	links, err := s.linkRepo.GetAll()
	if err != nil {
		return err
	}

	for {
		for _, link := range links {
			go s.crawlHub(link)

			time.Sleep(TIME_HUB_CRAWLS)
		}

		time.Sleep(CRAWL_ITERATION_TIME)
	}
}

func (s *Crawler) crawlHub(link model.Link) {

	doc, err := s.makeRequest(link.URL)
	if err != nil {
		log.Println("makeRequest:", err)
	}

	node, err := html.Parse(strings.NewReader(doc))
	if err != nil {
		log.Println("html.Parse:", err)
	}

	hubInfo := s.findHubInfo(link, node)
	hubInfo, err = s.hubRepo.Create(hubInfo)
	if err != nil || hubInfo.ID == 0 {
		return
	}
	s.printModel("HUB", hubInfo)

	urls := s.findArticleUrls(node, false, false)

	articleNodes := make(map[string]*html.Node)
	reqCh := make(chan RequestResult)
	
	nodesCount := 0
	for _, url := range urls {
		go s.makeRequestAsync(url, reqCh)

		time.Sleep(TIME_BETWEEN_REQUESTS)
	}

	for i := 0; i < len(urls); i++ {
		result := <-reqCh
		if result.IsErr {
			continue
		}
		
		node, err := html.Parse(strings.NewReader(result.Doc))
		if err != nil {
			log.Println("html.Parse:", err)
			continue
		}

		articleNodes[result.URL] = node
		nodesCount++
	}
	close(reqCh)

	infoCh := make(chan ArticleInfoResult)
	
	for url, node := range articleNodes {
		go s.findArticleInfo(hubInfo, url, node, infoCh)
	}

	var isErrCount int
	var repoErr int

	for i := 0; i < nodesCount; i++ {
		info := <- infoCh
		if info.IsErr {
			isErrCount++
			continue
		}

		err := s.articleRepo.Create(info.Info)
		if err != nil {
			repoErr++
			continue
		}

		s.printModel("ARTICLE", info.Info)
	}

	fmt.Printf("%d : %d : %d\n", hubInfo.ID, isErrCount, repoErr)
}

func (s *Crawler) printModel(modelName string, model any) {
	modelType := reflect.TypeOf(model)
	modelValue := reflect.ValueOf(model)

	var builder strings.Builder

	builder.WriteString("-------------------------------\n")
	builder.WriteString(fmt.Sprintf("%s\n", modelName))
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		fieldValue := modelValue.Field(i)

		if (field.Type.Name() == "Model" && field.PkgPath == "gorm") || !s.fieldNotModel(field.Name) {
			continue
		}

		builder.WriteString(fmt.Sprintf("%s: %v\n", field.Name, fieldValue.Interface()))
	}
	builder.WriteString("-------------------------------\n\n")

	fmt.Println(builder.String())
	builder.Reset()
}

func (s *Crawler) fieldNotModel(field string) bool {
	models := []string{"Model", "Link", "Hub", "Article"}

	for _, m := range models {
		if field == m {

			return false
		}
	}

	return true
} 

func (s *Crawler) makeRequest(url string) (string, error) {
	response, err := s.httpClient.Get(url)
	if err != nil {
		return "", nil
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (s *Crawler) makeRequestAsync(url string, ch chan<- RequestResult) {

	doc, err := s.makeRequest(url)
	ch <- RequestResult{
		URL: url,
		Doc: doc,
		IsErr: err != nil,
	}
}

func (s *Crawler) getAttrValue(node *html.Node, attrName string) string {

	var url string
	for _, attr := range node.Attr {
		if attr.Key == attrName {
			url = attr.Val
		}
	}

	return url
}

func (s *Crawler) findNodeByTag(node *html.Node, tag string) *html.Node {
	if node.Type == html.ElementNode && node.Data == tag {
		return node
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		foundNode := s.findNodeByTag(c, tag)
		if foundNode != nil {
			return foundNode
		}
	}

	return nil
}

func (s *Crawler) checkNodeHasClass(node *html.Node, class string) bool {
	for _, attr := range node.Attr {
		if attr.Key == "class" && strings.Contains(attr.Val, class) {
			return true
		}
	}

	return false
}

func (s *Crawler) findHubInfo(link model.Link, node *html.Node) model.Hub {
	info := model.Hub{
		LinkID: link.ID,
	}

	var crawler func(*html.Node)
	crawler = func (n *html.Node)  {
		if n.Type == html.ElementNode {
			if n.DataAtom == atom.H1 {
				info.Title = n.FirstChild.FirstChild.Data
			} else if n.DataAtom == atom.P && s.checkNodeHasClass(n, "tm-hub-card__description") {
				info.Description = n.FirstChild.Data
			} else if n.DataAtom == atom.Span && s.checkNodeHasClass(n, "tm-votes-lever__score-counter") {
				rating, _ := strconv.ParseFloat(strings.TrimSpace(n.FirstChild.Data), 64)
				info.Rating = rating
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			crawler(c)
		}
	}

	crawler(node)

	return info
}

func (s *Crawler) findArticleUrls(node *html.Node, isArticle, isH2 bool) []string {
	var urls []string

	var builder strings.Builder
	builder.Grow(ARTICLE_URL_LEN)

	// O(N)
	var crawler func(*html.Node, bool, bool)
	crawler = func(n *html.Node, isArticle, isH2 bool) {
		if n.Type == html.ElementNode {
			if n.DataAtom == atom.Article {
				isArticle = true
			} else if isArticle && !isH2 && n.DataAtom == atom.H2 {
				isH2 = true
			} else if isH2 && n.DataAtom == atom.A {
				articlePath := s.getAttrValue(n, "href")

				builder.WriteString(BASE_URL)
				builder.WriteString(articlePath)

				urls = append(urls, builder.String())

				builder.Reset()
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			crawler(c, isArticle, isH2)
		}
	}

	crawler(node, false, false)

	return urls
}

func (s *Crawler) findArticleInfo(hub model.Hub, url string, node *html.Node, ch chan<- ArticleInfoResult) {
	info := model.Article{
		HubID: hub.ID,
		URL: url,
	}
	
	articleNode := s.findNodeByTag(node, "article")
	if articleNode == nil {
		ch <- ArticleInfoResult{
			Info: info,
			IsErr: true,
		}
		return
	}

	var crawler func(*html.Node, bool)
	crawler = func(n *html.Node, isH1 bool) {
		if n.Type == html.ElementNode {
			if !isH1 && n.DataAtom == atom.H1 {
				isH1 = true
			} else if isH1 && n.DataAtom == atom.Span {
				info.Title = n.FirstChild.Data
			} else if n.DataAtom == atom.A && s.checkNodeHasClass(n, "tm-user-info__username") {
				userSuf := s.getAttrValue(n, "href")

				var builder strings.Builder
				builder.Grow(len(BASE_URL) + len(userSuf))
				builder.WriteString(BASE_URL)
				builder.WriteString(userSuf)

				info.AuthorLink = builder.String()
				info.Author = strings.TrimSpace(n.FirstChild.Data)
			} else if n.DataAtom == atom.Time {
				info.Datetime = s.getAttrValue(n, "datetime")
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			crawler(c, isH1)
		}
	}

	crawler(articleNode.FirstChild, false)

	ch <- ArticleInfoResult{
		Info: info,
		IsErr: false,
	}
}