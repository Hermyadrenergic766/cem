# ⚡ CEM — Unified AI Orchestrator · [cem.pw](https://cem.pw)

```
   ██████╗███████╗███╗   ███╗
  ██╔════╝██╔════╝████╗ ████║
  ██║     █████╗  ██╔████╔██║
  ██║     ██╔══╝  ██║╚██╔╝██║
  ╚██████╗███████╗██║  ╚═╝ ██║
   ╚═════╝╚══════╝╚═╝      ╚═╝
```

Birden fazla AI CLI'ı tek komutla kullan.
Claude düşünsün, Agy yazsın — ya da istediğin kombinasyon.

---

## Kur

```sh
# macOS & Linux
curl -fsSL cem.pw/install | sh

# Windows (PowerShell)
irm cem.pw/install.ps1 | iex
```

İlk `cem` komutunda wizard açılır.

---

## Kullan

```sh
cem "soru"           # thinker AI — varsayılan, think yazmaya gerek yok
cem -w "görev"       # writer AI
cem -p "görev"       # pair: düşün → yaz
cem -f dosya.py      # dosyayı thinker'a gönder
cem -wf dosya.py     # dosyayı writer'a gönder
cat kod.py | cem -p  # pipe ile pair
```

## AI Araçları

```sh
cemi           # kurulu araçları listele
cemi claude    # Claude kur
cemi agy       # Agy kur
cemi all       # hepsini kur (onay sorarak)
cemi update    # hepsini güncelle
cemir claude   # Claude kaldır
cemir agy      # Agy kaldır
```

## Roller

```sh
cem roles                    # kim ne yapıyor?
cem roles claude agy         # global: thinker=claude, writer=agy
cem roles gemini             # sadece thinker değiştir
cem roles --here claude agy  # sadece bu proje
cem init                     # proje wizard (interaktif)
cem init claude agy          # proje config direkt oluştur
```

---

## Config Dosyaları

| Dosya | Kapsam |
|-------|--------|
| `~/.cem/config.yaml` | Global — tüm projeler |
| `./.cem.yaml` | Proje — sadece bu dizin |

Proje dizininde `.cem.yaml` varsa global'i override eder.
Yoksa global kullanılır ve `cem` her komutta bunu bildirir.

---

## Desteklenen Araçlar

| Araç | Açıklama |
|------|----------|
| `claude` | Analiz, mimari, düşünme |
| `agy` | Hızlı kod üretimi |
| `aider` | Git-aware kod yazma |
| `gemini` | Google Gemini |
| `gpt` | OpenAI GPT-4 |
