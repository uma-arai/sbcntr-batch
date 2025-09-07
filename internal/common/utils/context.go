package utils

import (
	"context"
	"fmt"
	"time"
)

// 指定されたタイムアウト時間内でバッチ処理を実行する
// タイムアウトを超えた場合は、コンテキストをキャンセルしてエラーを返す
func RunWithTimeout(ctx context.Context, timeout time.Duration, fn func(context.Context) error) error {
	// タイムアウト付きのコンテキストを作成
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// エラーチャネルを作成
	errChan := make(chan error, 1)

	// バッチ処理を実行
	go func() {
		errChan <- fn(ctx)
	}()

	// バッチ処理の完了またはタイムアウトを待機
	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return fmt.Errorf("batch process timed out after %v", timeout)
	}
}
