package searcher

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode"
	"word-search-in-files/pkg/internal/dir"
)

type Searcher struct {
	FS fs.FS
}

func (s *Searcher) Search(word string) ([]string, error) {
	fileNames, err := dir.FilesFS(s.FS, ".")
	if err != nil {
		return nil, err
	}

	var (
		mutex   sync.Mutex
		wg      sync.WaitGroup
		results []string
		errors  []error
	)
//запускаем отдельную горутину для поиска по каждому файлу
	for _, fileName := range fileNames {
		wg.Add(1)
		go func(fileName string) {
			defer wg.Done()

			content, err := fs.ReadFile(s.FS, fileName)
			if err != nil {
				mutex.Lock()
				errors = append(errors, err)
				mutex.Unlock()
				return
			}

			//разделяем слова пробелами и удаляем знаки препинания
			words := strings.FieldsFunc(string(content), func(r rune) bool {
				return unicode.IsPunct(r) || unicode.IsSpace(r)
			})

			//используем поиск слов без учета регистра
			for _, w := range words {
				if strings.EqualFold(w, word) {
					mutex.Lock()
					results = append(results, fileName)
					mutex.Unlock()
					break
				}
			}
		}(fileName)
	}

	wg.Wait()

	if len(errors) > 0 {
		return nil, errors[0]
	}

	return results, nil
}

func SearchFiles(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Получаем ключевое слово из адресной строки
	parts := strings.Split(path, "/")
	word := parts[len(parts)-1]

	if word == "" {
		http.Error(w, "Необходимо указать ключевое слово", http.StatusBadRequest)
		return
	}

	directory := os.DirFS("./examples")
	searcher := &Searcher{FS: directory}
	files, err := searcher.Search(word)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Удаляем расширения файлов из результатов
	for i, fileName := range files {
		files[i] = strings.TrimSuffix(fileName, filepath.Ext(fileName))
	}

	response := struct {
		Files []string `json:"files"`
	}{
		Files: files,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
