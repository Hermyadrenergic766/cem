---
description: Görevi cem'e devret — thinker ve writer farklı AI'lar olur (token tasarrufu)
---

# /cem — multi-AI pair çağrısı

Kullanıcının görevi cem'e pair modunda devredilir. **Mevcut session'da çalışan AI (sen) bu görevi yapmazsın** — sadece komutu çalıştırıp çıktıyı yansıtırsın. Asıl iş, `~/.cem/config.yaml`'da kayıtlı **thinker** ve **writer** AI'larında yapılır.

**Token tasarrufu mantığı:**
- Sen (host AI) sadece pass-through olduğun için minimum token harcarsın
- Thinker (örn. claude/haiku — ucuz model) analiz/planlama yapar
- Writer (örn. agy/gemini-3-flash — başka provider) kodu yazar
- cem akıllı skip kurallarıyla gereksiz writer çağrısını da atlar

Bu sayede pahalı bir "büyük model"i hem analiz hem yazım için ayrı ayrı harcamak yerine, görevi iki farklı (uygun) modele dağıtırsın.

**Yapman gereken:**
1. Şu komutu çalıştır (Bash tool):

   ```bash
   cem -p "$ARGUMENTS"
   ```

2. Çıktıyı kullanıcıya **olduğu gibi** göster. Yeniden yorumlama, yeniden analiz etme, kendi sürümünü yazma.
3. Komut hata verirse `cem doctor` çalıştırmasını öner.

**Kullanıcının görevi:** $ARGUMENTS

---

> Roller ve modeller `~/.cem/config.yaml`'da. Değiştirmek için:
> - `cem setup` (interaktif wizard)
> - `cem roles <thinker> <writer>`
> - IDE plugin Settings → Tools → cem
