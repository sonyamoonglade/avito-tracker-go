package parser

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
)

// Uses chrome dev tools protocol to parse data
type ChromeParser struct {
	matchers [4]*regexp.Regexp
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewChromeParser() (*ChromeParser, error) {

	titlerx := regexp.MustCompile(`"title-info-title-text.*?<`)
	pricerx := regexp.MustCompile(`js-item-price.*?<`)
	imagerx1 := regexp.MustCompile("image-frame-cover.*?<div")
	imagerx2 := regexp.MustCompile(`src="(.*?)"`)

	// Start browser instance
	ctx, cancel := chromedp.NewContext(context.Background())
	if err := chromedp.Run(ctx); err != nil {
		return nil, fmt.Errorf("error booting a browser: %w", err)
	}

	// TODO: remove panic
	go func() {
		<-chromedp.FromContext(ctx).Browser.LostConnection
		panic("browser is down")
	}()

	return &ChromeParser{
		matchers: [4]*regexp.Regexp{titlerx, pricerx, imagerx1, imagerx2},
		ctx:      ctx,
		cancel:   cancel,
	}, nil
}

func (p *ChromeParser) Parse(timeout time.Duration, url string) *ParseResult {
	var html string

	// get rid of somehow
	ctx, cancel := context.WithTimeout(p.ctx, timeout)
	defer cancel()
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.ActionFunc(func(c context.Context) error {
			node, err := dom.GetDocument().Do(c)
			if err != nil {
				return fmt.Errorf("parser: could not get document: %w", err)
			}

			res, err := dom.GetOuterHTML().WithNodeID(node.NodeID).Do(c)
			if err != nil {
				return fmt.Errorf("parser: could not get outer html of body: %w", err)
			}

			html = res
			return nil
		}),
	)

	title, err := p.parseTitle(&html)
	if err != nil {
		err = fmt.Errorf("parser: title parsing error: %w", err)
		return NewParseResultWithError(err)
	}

	price, err := p.parsePrice(&html)
	if err != nil {
		err = fmt.Errorf("parser: price parsing error: %w", err)
		return NewParseResultWithError(err)
	}

	return NewParseResult(title, price, url)
}

func (p *ChromeParser) parseTitle(buff *string) (string, error) {

	rx := p.matchers[0]
	results := rx.FindAllString(*buff, 1)

	// URL is unavailable or ip is banned
	if len(results) == 0 {
		return "", ErrURLUnavailable
	}

	spl := strings.Split(results[0], "")

	var title string
	for i := len(spl) - 2; i >= 0; i-- {
		if spl[i] == ">" {
			break
		}

		title = spl[i] + title
	}

	return title, nil
}

func (p *ChromeParser) parsePrice(buff *string) (float64, error) {

	rx := p.matchers[1]
	results := rx.FindAllString(*buff, 1)

	if len(results) == 0 {
		return 0.0, ErrURLUnavailable
	}

	spl := strings.Split(results[0], "")

	var pricestr string
	for i := len(spl) - 2; i >= 0; i-- {
		if spl[i] == ">" {
			break
		}

		// Compare by charcode (leave only numbers)
		if spl[i][0] < 48 || spl[i][0] > 57 {
			continue
		}

		pricestr = spl[i] + pricestr
	}

	pricefloat, err := strconv.ParseFloat(pricestr, 64)
	if err != nil {
		return 0.0, err
	}

	return pricefloat, nil
}
