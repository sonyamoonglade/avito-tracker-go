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

func NewChromeParser(timeout time.Duration) (*ChromeParser, error) {

	titlerx, err := regexp.Compile(`"title-info-title-text.*?<`)
	if err != nil {
		return nil, err
	}

	pricerx, err := regexp.Compile(`js-item-price.*?<`)
	if err != nil {
		return nil, err
	}

	imagerx1, err := regexp.Compile("image-frame-cover.*?<div")
	if err != nil {
		return nil, err
	}
	imagerx2, err := regexp.Compile(`src="(.*?)"`)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	ctx, cancel = chromedp.NewContext(ctx)
	return &ChromeParser{
		matchers: [4]*regexp.Regexp{titlerx, pricerx, imagerx1, imagerx2},
		ctx:      ctx,
		cancel:   cancel,
	}, nil
}

func (p *ChromeParser) Parse(url string) (*ParseResult, error) {
	var html string
	err := chromedp.Run(p.ctx,
		chromedp.Navigate(url),
		chromedp.ActionFunc(func(c context.Context) error {
			node, err := dom.GetDocument().Do(c)
			if err != nil {
				return fmt.Errorf("could not get document: %w", err)
			}

			res, err := dom.GetOuterHTML().WithNodeID(node.NodeID).Do(c)
			if err != nil {
				return fmt.Errorf("could not get outer html of body: %w", err)
			}

			html = res
			return nil
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("parsing error: %w", err)
	}

	title := p.parseTitle(&html)
	price, err := p.parsePrice(&html)
	if err != nil {
		return nil, fmt.Errorf("price parsing error: %w", err)
	}

	return &ParseResult{
		Title: title,
		Price: price,
	}, nil
}

func (p *ChromeParser) parseTitle(buff *string) string {
	rx := p.matchers[0]

	match := rx.FindAllString(*buff, 1)[0]
	spl := strings.Split(match, "")

	var title string
	for i := len(spl) - 2; i >= 0; i-- {
		if spl[i] == ">" {
			break
		}

		title = spl[i] + title
	}

	return title
}

func (p *ChromeParser) parsePrice(buff *string) (float64, error) {
	rx := p.matchers[1]

	match := rx.FindAllString(*buff, 1)[0]
	spl := strings.Split(match, "")

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
