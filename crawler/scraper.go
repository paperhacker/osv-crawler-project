package crawler

import (
    "bytes"
    "context"
    "fmt"
    "net/http"
    "net/url"
    "strings"
    "time"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/service/s3"
    "github.com/aws/aws-sdk-go-v2/service/s3/types"
    "github.com/gocolly/colly/v2"
    "github.com/rs/zerolog/log"
    "github.com/paperhacker/osv-crawler-project/metrics"
    "github.com/paperhacker/osv-crawler-project/input"
)

func Scrape(ctx context.Context, target input.Target, bucket, prefix string, s3Client *s3.Client) error {
    c := colly.NewCollector(colly.Async(true))
    c.WithTransport(&http.Transport{Proxy: http.ProxyFromEnvironment})
    c.SetRequestTimeout(20 * time.Second)

    var scrapeErr error

    c.OnResponse(func(r *colly.Response) {
        parsed, _ := url.Parse(r.Request.URL.String())
        now := time.Now().Format("2006-01-02")
        filename := sanitize(target.URL) + ".html"
        s3Key := fmt.Sprintf("%s%s/%s/%s", strings.TrimSuffix(prefix, "/"), parsed.Host, now, filename)

        _, err := s3Client.PutObject(ctx, &s3.PutObjectInput{
            Bucket:      aws.String(bucket),
            Key:         aws.String(s3Key),
            Body:        bytes.NewReader(r.Body),
            ContentType: aws.String("text/html"),
            ACL:         types.ObjectCannedACLPrivate,
        })
        if err != nil {
            scrapeErr = fmt.Errorf("failed to upload: %w", err)
            log.Error().Err(err).Str("key", s3Key).Msg("Upload failed")
            metrics.CrawlFailures.WithLabelValues(target.Name).Inc()
        } else {
            log.Info().Str("key", s3Key).Msg("Upload successful")
            metrics.PagesCrawled.WithLabelValues(target.Name).Inc()
        }
    })

    c.OnError(func(r *colly.Response, err error) {
        scrapeErr = fmt.Errorf("scrape error: %w", err)
        metrics.CrawlFailures.WithLabelValues(target.Name).Inc()
    })

    if err := c.Visit(target.URL); err != nil {
        return fmt.Errorf("visit now failed: %w", err)
    }

    c.Wait()
    return scrapeErr
}

func sanitize(url string) string {
    return strings.NewReplacer("https://", "", "http://", "", "/", "_", "?", "_", ":", "_").Replace(url)
}
