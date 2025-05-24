package input

import (
    "context"
    "encoding/json"
    "fmt"
    "strings"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/service/s3"
)

type Target struct {
    URL  string `json:"url"`
    Name string `json:"name"`
}

func LoadTargets(ctx context.Context, s3Client *s3.Client, uri string) ([]Target, error) {
    parts := strings.SplitN(strings.TrimPrefix(uri, "s3://"), "/", 2)
    if len(parts) != 2 {
        return nil, fmt.Errorf("invalid S3 URI: %s", uri)
    }

    bucket, key := parts[0], parts[1]
    out, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
        Bucket: aws.String(bucket),
        Key:    aws.String(key),
    })
    if err != nil {
        return nil, fmt.Errorf("failed to fetch targets file: %w", err)
    }
    defer out.Body.Close()

    var targets []Target
    if err := json.NewDecoder(out.Body).Decode(&targets); err != nil {
        return nil, fmt.Errorf("failed to decode targets: %w", err)
    }

    return targets, nil
}
