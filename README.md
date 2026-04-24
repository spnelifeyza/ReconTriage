# ReconTriage

<p align="center">
  <img src="images/logo.png" width="600"/>
</p>

<p align="center">
  <b>Reconnoiter • Analyze • Prioritize</b>
</p>

---

## 🚀 About

ReconTriage is an automated reconnaissance and analysis tool designed for security researchers and bug bounty hunters.

Instead of overwhelming users with raw recon data, it:
- Collects subdomains automatically
- Processes and analyzes results
- Prioritizes high-value targets

👉 Focus on what matters. Skip the noise.

---

## ⚙️ Requirements

Make sure the following tools are installed and accessible in your system:

- subfinder  
- assetfinder  

---

### 🔍 Install subfinder

```
go install github.com/projectdiscovery/subfinder/v2/cmd/subfinder@latest
```

---

### 🔎 Install assetfinder

```
go install github.com/tomnomnom/assetfinder@latest
```

---

## ▶️ Usage

```
.\bin\recontriage.exe example.com
```

---

## 🧠 How It Works

1. Collect subdomains using subfinder & assetfinder  
2. Merge and remove duplicates  
3. Analyze targets (title extraction + keyword matching)  
4. Assign severity levels  
5. Output prioritized results  

---

## 🏗️ Project Structure

```
ReconTriage/
│── bin/
│   └── recontriage.exe
│
│── configs/
│   └── keywords.json
│
│── images/
│   └── logo.png
│
│── internal/
│   ├── analyzer/
│   ├── host/
│   ├── subdomain/
│   └── utils/
│
│── outputs/
│── README.md
```

---

## 📊 Output

Results are saved in:

```
outputs/all.txt
outputs/alive.txt
outputs/timeout.txt
outputs/errors.txt
outputs/results.json
outputs/report.txt
```

### 📁 File Descriptions

- **alive.txt**  
  Targets that responded successfully (HTTP response received).

- **all.txt**  
  Contains all discovered subdomains collected from recon tools.

- **errors.txt**  
  Targets that returned connection or request errors during scanning.

- **report.txt**  
  Human-readable summary of findings with prioritized targets.

- **results.json**  
  Final analyzed output including URL, page title, and severity level.

- **timeout.txt**  
  Targets that did not respond within the specified timeout duration.
---

## 🎯 Roadmap

- [ ] HTTP probing (alive hosts)  
- [ ] CLI improvements  
- [ ] Auto-report generation  
- [ ] Colored terminal output  

---

Developed by **Feyza SAPAN**

