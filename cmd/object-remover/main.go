package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	remover "github.com/minikocha/kiro-university-challenge"
)

type flags struct {
	bucket    string
	prefix    string
	region    string
	yes       bool
	help      bool
	usageFunc func()
}

// parseFlags は引数を解析して *flags と error を返す。
// fs.Parse が失敗した場合は &f と error を返す。
// --bucket または --prefix が未指定の場合はヘルプを表示して異常終了する。
func parseFlags(args []string) (*flags, error) {
	var f flags

	fs := flag.NewFlagSet("object-remover", flag.ContinueOnError)
	fs.StringVar(&f.bucket, "bucket", "", "S3 bucket name (required)")
	fs.StringVar(&f.prefix, "prefix", "", "prefix of the objects to delete (required)")
	fs.StringVar(&f.region, "region", "", "AWS region of the S3 bucket (defaults to AWS_REGION env or ~/.aws/config)")
	fs.BoolVar(&f.yes, "yes", false, "skip yes/no confirmation before deleting")
	fs.BoolVar(&f.help, "help", false, "show help and exit")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: object-remover --bucket <bucket> --prefix <prefix> [--region <region>] [--yes] [--help]\n\n")
		fmt.Fprintf(os.Stderr, "Deletes all objects (including all versions and delete markers) under the specified prefix in an S3 bucket.\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		fs.PrintDefaults()
	}
	f.usageFunc = fs.Usage

	if err := fs.Parse(args); err != nil {
		return &f, err
	}

	if f.help {
		return &f, nil
	}

	if f.bucket == "" || f.prefix == "" {
		return &f, fmt.Errorf("--bucket and --prefix are required")
	}

	return &f, nil
}

func main() {
	f, err := parseFlags(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n\n", err)
		f.usageFunc()
		os.Exit(1)
	}

	if f.help {
		f.usageFunc()
		os.Exit(0)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT)
	defer stop()

	// S3クライアント初期化
	r, err := remover.NewRemover(ctx, f.bucket, f.prefix, f.region)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// バケット存在確認
	if err := r.CheckBucketExists(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// オブジェクト列挙
	objects, err := r.ListObjects(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if len(objects) == 0 {
		fmt.Printf("No objects found with prefix %q in bucket %q.\n", f.prefix, f.bucket)
		os.Exit(0)
	}

	// --yes 未指定時は一覧表示して確認
	if !f.yes {
		fmt.Printf("The following %d object(s) will be deleted:\n", len(objects))
		for _, obj := range objects {
			fmt.Printf("  s3://%s/%s (versionId: %s)\n", f.bucket, obj.Key, obj.VersionId)
		}
		fmt.Print("\nDo you want to delete these objects? [Yes/No]: ")

		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		answer := strings.TrimSpace(scanner.Text())

		if !strings.EqualFold(answer, "yes") {
			fmt.Println("Cancelled.")
			os.Exit(0)
		}
	}

	// 削除実施
	failed, err := r.DeleteObjects(ctx, objects)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if len(failed) > 0 {
		fmt.Fprintf(os.Stderr, "Failed to delete %d object(s):\n", len(failed))
		for _, obj := range failed {
			fmt.Fprintf(os.Stderr, "  s3://%s/%s (versionId: %s)\n", f.bucket, obj.Key, obj.VersionId)
		}
		os.Exit(1)
	}

	fmt.Printf("Successfully deleted %d object(s).\n", len(objects))
}
