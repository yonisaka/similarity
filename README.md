## How to use

### 1. Set up OpenAI API Key on 

```go
    _ = os.Setenv("OPENAI_API_KEY", "test") // your API key
```

### 2. Run *search_test* with your prompt and expected result


```go
    answer, err := searchUsecase.Search(ctx, "stok nomor BA00001323K14 memiliki plat nomor apa?")
    if err != nil {
        log.Println(err)
        assert.Error(t, err)
        return
    }
    
    assert.Contains(t, answer, "B1207KDZ")
    log.Println(answer)
```
