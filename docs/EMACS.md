# cem in Emacs

No plugin needed — cem is a shell command.

---

## Basic recipe

Add to `init.el` / `.emacs`:

```elisp
(defun cem-region (mode start end)
  "Send region to cem (think/write/pair) and show output in a buffer."
  (interactive (list (completing-read "Mode: " '("think" "write" "pair"))
                     (region-beginning) (region-end)))
  (let* ((prompt (buffer-substring-no-properties start end))
         (flag (pcase mode
                 ("think" "")
                 ("write" "-w")
                 ("pair" "-p")))
         (cmd (concat "cem " flag " "
                      (shell-quote-argument prompt)))
         (buf (get-buffer-create "*cem*")))
    (with-current-buffer buf
      (erase-buffer)
      (insert (format "─── cem %s ───\n" mode)))
    (start-process-shell-command "cem" buf cmd)
    (display-buffer buf)))

(defun cem-think () (interactive)
  (cem-region "think" (region-beginning) (region-end)))
(defun cem-write () (interactive)
  (cem-region "write" (region-beginning) (region-end)))
(defun cem-pair () (interactive)
  (cem-region "pair" (region-beginning) (region-end)))

(global-set-key (kbd "C-c c t") 'cem-think)
(global-set-key (kbd "C-c c w") 'cem-write)
(global-set-key (kbd "C-c c p") 'cem-pair)
```

Select a region, then `C-c c t/w/p`. Output streams into `*cem*` buffer.

---

## With Doom Emacs

Add to `~/.doom.d/config.el`:

```elisp
(map! :leader
      (:prefix-map ("c" . "cem")
       :desc "Think on region" "t" #'cem-think
       :desc "Write on region" "w" #'cem-write
       :desc "Pair on region"  "p" #'cem-pair))
```

Now `SPC c t/w/p` invokes cem.

---

## Org-mode integration

```elisp
(defun cem-on-src-block (mode)
  "Run cem on the current org-mode src block contents."
  (interactive (list (completing-read "Mode: " '("think" "write" "pair"))))
  (when-let* ((elem (org-element-context))
              (info (org-element-property :value elem)))
    (cem-region mode (point-min) (point-min))
    (with-current-buffer "*cem*"
      (insert info))))
```

Place point in a src block, `M-x cem-on-src-block`.
