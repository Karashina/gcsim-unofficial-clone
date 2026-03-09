package assets

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/go-chi/chi"
)

func (s *Server) handleGetData(t AssetType) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := chi.URLParam(r, "key")
		s.logger.Info("request for data", "t", t.String(), "key", key)

		// 特殊アセットは別途処理する
		if _, ok := specialAssets[key]; ok {
			s.logger.Info("key is special", "key", key)
			data, err := staticAssets.ReadFile(fmt.Sprintf("static/special/%v.png", key))
			if err != nil {
				s.logger.Warn("unexpected special key but asset not found", "key", key, "err", err)
				s.handleNotFound(w)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write(data)
			return
		}

		// 認識できるアイテムか確認。認識できなければ残りの処理をスキップ
		assetName, ok := s.assetNameMapping[t][key]
		if !ok {
			s.logger.Info("unrecognized key, serving default", "t", t.String(), "key", key)
			s.handleNotFound(w)
			return
		}

		// まずキャッシュを試す
		data, err := s.loadFromCache(t, key)
		switch {
		case data != nil && err == nil:
			// 読み込み成功
			s.logger.Info("loaded from cache ok", "t", t.String(), "key", key)
			w.WriteHeader(http.StatusOK)
			w.Write(data)
			return
		case err != nil:
			s.logger.Info("unexpected error trying to read from cache", "t", t.String(), "key", key)
		case data == nil:
			s.logger.Info("cache data not found", "t", t.String(), "key", key)
		}

		// 外部ソースを1つずつ試す
		for i, v := range s.hosts[t] {
			joinedURL := v.JoinPath(fmt.Sprintf("/%v.png", assetName))
			s.logger.Info("trying external image source", "host", v.String(), "try", i, "key", key, "full_path", joinedURL.String())
			data, err := s.proxyImageRequest(joinedURL)
			if err != nil {
				s.logger.Info("error getting", "err", err, "host", v.String(), "try", i, "key", key, "full_path", joinedURL.String())
				continue
			}
			s.logger.Info("received image from source ok", "host", v.String())
			s.saveToCache(t, key, data)
			// 見つかったのでキャッシュに保存してリクエストを終了
			w.WriteHeader(http.StatusOK)
			w.Write(data)
			return
		}
		s.logger.Info("no external source found, serving default", "t", t.String(), "key", key)
		// ここに到達した場合、全ての外部ソースが失敗したのでデフォルトのクエスチョンマーク画像を配信
		s.handleNotFound(w)
	}
}

func (s *Server) handleNotFound(w http.ResponseWriter) {
	// /static/misc/default.png を返すべき
	data, err := staticAssets.ReadFile("static/misc/default.png")
	if err != nil {
		s.logger.Warn("error reading default.png", "err", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (s *Server) saveToCache(t AssetType, key string, data []byte) {
	fp := path.Join(s.cacheDir, fmt.Sprintf("/%v/%v.png", t.String(), key))
	f, err := os.Create(fp)
	if err != nil {
		s.logger.Warn("error writing to cache", "err", err)
	}
	_, err = f.Write(data)
	if err != nil {
		s.logger.Warn("error writing to cache", "err", err)
	}
	f.Close()
}

func (s *Server) loadFromCache(t AssetType, key string) ([]byte, error) {
	fp := path.Join(s.cacheDir, fmt.Sprintf("/%v/%v.png", t.String(), key))
	_, err := os.Stat(fp)
	if err == nil {
		return os.ReadFile(fp)
	} else if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	return nil, fmt.Errorf("unexpected error checking file: %w", err)
}

func (s *Server) proxyImageRequest(path *url.URL) ([]byte, error) {
	resp, err := s.httpClient.Get(path.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return nil, fmt.Errorf("response is not an image: %v", contentType)
	}

	return io.ReadAll(resp.Body)
}

func (s *Server) handleOnlineCheck() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}
