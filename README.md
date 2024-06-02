# WappCaidoLayzer

Wrapper for the [wappalyzergo](https://github.com/projectdiscovery/wappalyzergo) library to be used in [Caido Workflows](https://github.com/caido/workflows).

# Example Usage

**In PowerShell Node:**
```powershell
$rr = $input | Out-String
$result = $rr | <Path to the exe>\.\wappcaidolayzer.exe -output "{{join .Tech.Categories}} -> {{.Name}}\n{{.Tech.Description}}\n {{.Tech.Website}}\n{{.Tech.Icon}}\n{{.Tech.CPE}}\n-----------------------------\n"
echo $result
```

**In Bash/Zsh/Shell Node:**
```bash
./wappcaidolayzer -output "{{join .Tech.Categories}} -> {{.Name}}\n{{.Tech.Description}}\n {{.Tech.Website}}\n{{.Tech.Icon}}\n{{.Tech.CPE}}\n-----------------------------\n"
```

**Note:** You can run it without `-output` or you can customize the output format.