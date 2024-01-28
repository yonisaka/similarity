## How to use

### 1. Set up OpenAI API Key on 

```go
    _ = os.Setenv("OPENAI_API_KEY", "test") // your API key
```

### 2. Set your question and answer on


```go
    qna := make(map[string]string, 4)
    qna["stok nomor BA00002123J16 dimiliki penjual apa?"] = "Yuliana adec "
    qna["stok nomor BA00001323K14 memiliki plat nomor apa?"] = "B1207KDZ"
    qna["mobil dengan plat nomor F1088DA memiliki warna apa?"] = "Hitam Metalic"
    qna["mobil dengan plat nomor B1690PRD memiliki harga awal berapa?"] = "62000000"
```

### 3. Run *search_test*