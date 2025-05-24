package main

import (
    "context"
    "flag"
    "os"
    "strings"
    "sync"

    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/s3"
    "github.com/joho/godotenv"
    "github.com/rs/zerolog/log"

    "github.com/paperhacker/osv-crawler-project/config"
    "github.com/paperhacker/osv-crawler-project/crawler"
    "github.com/paperhacker/osv-crawler-project/input"
    "github.com/paperhacker/osv-crawler-project/logger"
    "github.com/paperhacker/osv-crawler-project/metrics"
)

func main() {
    envFile := flag.String("env-file", "", "Path to .env file")
    flag.Parse()

    if *envFile != "" {
        _ = godotenv.Load(*envFile)
    }

    logger.Init()

    cfg := config.Load()

    ctx := context.Background()
    awsCfg, err := config.LoadDefaultConfig(ctx)
    if err != nil {
        log.Fatal().Err(err).Msg("Unable to load AWS config")
    }
    s3Client := s3.NewFromConfig(awsCfg)

    go metrics.Serve()

    targets, err := input.LoadTargets(ctx, s3Client, cfg.TargetsS3Uri)
    if err != nil {
        log.Fatal().Err(err).Msg("Failed to load targets")
    }

    var wg sync.WaitGroup
    for _, target := range targets {
        wg.Add(1)
        go func(t input.Target) {
            defer wg.Done()
            if err := crawler.Scrape(ctx, t, cfg.OutputS3Bucket, cfg.OutputS3Prefix, s3Client); err != nil {
                log.Error().Err(err).Str("url", t.URL).Msg("Crawl failed")
            } else {
                log.Info().Str("url", t.URL).Msg("Crawl complete")
            }
        }(target)
    }

    wg.Wait()
    log.Info().Msg("All crawls complete")
}
