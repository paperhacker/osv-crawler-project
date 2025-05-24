package config

import "os"

type CrawlerConfig struct {
    TargetsS3Uri    string
    OutputS3Bucket  string
    OutputS3Prefix  string
    ProxyURL        string
    TaskTag         string
}

func Load() CrawlerConfig {
    return CrawlerConfig{
        TargetsS3Uri:   os.Getenv("TARGETS_S3_URI"),
        OutputS3Bucket: os.Getenv("OUTPUT_S3_BUCKET"),
        OutputS3Prefix: os.Getenv("OUTPUT_S3_PREFIX"),
        ProxyURL:       os.Getenv("PROXY_URL"),
        TaskTag:        os.Getenv("TASK_TAG"),
    }
}
