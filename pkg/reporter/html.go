// Keep your package and imports as is

package reporter

import (
	"fmt"
	"html/template"
	"os"
	"strings"

	"github.com/safedep/xbom/pkg/codeanalysis"
)

// HTMLVisualiser builds and writes an interactive HTML report
type HTMLVisualiser struct {
	headers []string
	rows    []map[string]interface{}
}

func NewHTMLVisualiser(headers []string) *HTMLVisualiser {
	return &HTMLVisualiser{
		headers: headers,
		rows:    []map[string]interface{}{},
	}
}

func (hv *HTMLVisualiser) AddRow(row map[string]interface{}) {
	hv.rows = append(hv.rows, row)
}

func (hv *HTMLVisualiser) Finish(htmlPath string) error {
	// Map of popular languages to their CDN icon links
	languageIconMap := map[string]string{
		"python":     "https://cdn.jsdelivr.net/gh/devicons/devicon/icons/python/python-original.svg",
		"javascript": "https://cdn.jsdelivr.net/gh/devicons/devicon/icons/javascript/javascript-original.svg",
		"go":         "https://cdn.jsdelivr.net/gh/devicons/devicon/icons/go/go-original.svg",
		"java":       "https://cdn.jsdelivr.net/gh/devicons/devicon/icons/java/java-original.svg",
		"c":          "https://cdn.jsdelivr.net/gh/devicons/devicon/icons/c/c-original.svg",
		"cpp":        "https://cdn.jsdelivr.net/gh/devicons/devicon/icons/cplusplus/cplusplus-original.svg",
		"ruby":       "https://cdn.jsdelivr.net/gh/devicons/devicon/icons/ruby/ruby-original.svg",
	}

	tmpl := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Signature Visualizer</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/choices.js/public/assets/styles/choices.min.css" />
    <script src="https://cdn.jsdelivr.net/npm/choices.js/public/assets/scripts/choices.min.js"></script>
</head>

<body>
<!-- Top Nav Bar -->
<nav class="bg-white border-b border-gray-200 px-4 py-3 flex items-center justify-between">
  <!-- Logo on the left -->
  <a href="https://safedep.io/" class="flex items-center">
  <img src="https://avatars.githubusercontent.com/u/115209633?s=200&v=4" alt="Logo" class="h-10 w-10 rounded-md mr-3">
  <span class="font-semibold text-lg text-gray-800">SafeDep</span>
</a>

  <!-- Star xbom on GitHub on the right -->
  <a href="https://github.com/safedep/xbom" target="_blank" rel="noopener noreferrer"
     class="flex items-center bg-gray-100 hover:bg-gray-200 text-gray-800 font-medium px-4 py-2 rounded transition">
    <!-- GitHub Logo SVG -->
    <svg class="h-5 w-5 mr-2" fill="currentColor" viewBox="0 0 24 24" aria-hidden="true">
      <path d="M12 0C5.37 0 0 5.373 0 12c0 5.303 3.438 9.8 8.205 11.387.6.113.82-.258.82-.577
      0-.285-.01-1.04-.015-2.04-3.338.726-4.042-1.61-4.042-1.61-.546-1.387-1.333-1.756-1.333-1.756-1.09-.745.083-.729.083-.729
      1.205.085 1.84 1.237 1.84 1.237 1.07 1.834 2.807 1.304 3.492.997.108-.775.418-1.305.76-1.605-2.665-.305-5.466-1.334-5.466-5.93
      0-1.31.468-2.38 1.236-3.22-.124-.303-.535-1.523.117-3.176 0 0 1.008-.322 3.3 1.23a11.52 11.52 0 013.003-.404c1.02.005
      2.047.138 3.003.404 2.29-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.873.12 3.176.77.84 1.235 1.91 1.235 3.22
      0 4.61-2.803 5.624-5.475 5.92.43.37.823 1.102.823 2.222 0 1.606-.014 2.898-.014 3.293 0 .322.216.694.825.576C20.565
      21.796 24 17.298 24 12c0-6.627-5.373-12-12-12z"/>
    </svg>
    Star us on GitHub
  </a>
</nav>

<div class="bg-gray-100 min-h-screen flex items-start justify-center p-6">
    <div class="bg-white p-8 rounded-2xl shadow-xl w-full max-w-6xl space-y-6 mt-[0] mb-[1%] sticky top-0">
        <h1 class="text-3xl font-bold mb-6 text-gray-800">Matched Signatures</h1>

        <!-- Filters -->
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
                <label class="block text-gray-700 mb-2" for="signatureSelect">Signature ID</label>
                <select id="signatureSelect" class="w-full border border-gray-300 rounded-lg p-2" multiple>
                    {{ range $index, $row := .Rows }}
                    <option value="{{ $row.Signature_ID }}">{{ $row.Signature_ID }}</option>
                    {{ end }}
                </select>
            </div>

            <div>
                <label class="block text-gray-700 mb-2" for="tagsSelect">Tags</label>
                <select id="tagsSelect" class="w-full border border-gray-300 rounded-lg p-2" multiple>
                    {{ range $tag := .UniqueTags }}
                        <option value="{{ $tag }}">{{ $tag }}</option>
                    {{ end }}
                </select>
            </div>
        </div>

        <!-- Table -->
        <div class="overflow-x-auto rounded-lg shadow bg-white">
            <table class="min-w-full divide-y divide-gray-200">
                <thead class="bg-gray-100">
                    <tr>
                        <th class="px-4 py-3 text-left text-xs font-semibold text-gray-700 uppercase tracking-wider">Signature ID</th>
                        <th class="px-4 py-3 text-left text-xs font-semibold text-gray-700 uppercase tracking-wider">Description</th>
                        <th class="px-4 py-3 text-left text-xs font-semibold text-gray-700 uppercase tracking-wider">Tags</th>
                        <th class="px-4 py-3 text-left text-xs font-semibold text-gray-700 uppercase tracking-wider">Details</th>
                    </tr>
                </thead>
                <tbody id="signatureTable" class="bg-white divide-y divide-gray-200">
                    {{ range $index, $row := .Rows }}
                    <tr data-signature="{{ $row.Signature_ID }}" data-tags="{{ $row.Tags }}">
                        <td class="px-4 py-2 text-sm text-gray-900">{{ $row.Signature_ID }}</td>
                        <td class="px-4 py-2 text-sm text-gray-900">{{ $row.Description }}</td>
                        <td class="px-4 py-2 text-sm text-gray-700">{{ $row.Tags }}</td>
                        <td class="px-4 py-2 text-sm">
                            <button class="flex items-center gap-1 text-blue-600 focus:outline-none font-medium px-2 py-1 rounded transition-colors duration-150 cursor-pointer border border-blue-200 bg-blue-50 hover:bg-blue-100"
                                onclick="toggleDetails('details-{{ $index }}', this)">
                                <span class="icon mr-1 transition-transform duration-200">&#9654;</span>
                                <span>Show Details</span>
                            </button>
                        </td>
                    </tr>
                    <tr id="details-{{ $index }}" class="details hidden">
                        <td colspan="4" class="bg-gray-50 px-4 py-2">
                            <div class="space-y-4">
                                {{ range $row.FileOccurrences }}
                                <div class="border rounded-lg p-4 bg-white shadow-md hover:shadow-lg transition-shadow duration-300">
                                    <div class="flex flex-wrap items-center gap-4 mb-2">
                                        <div class="text-xs text-gray-600"><strong>File:</strong> <span class="font-mono">{{ .File }}</span></div>
                                        <div class="flex items-center text-xs text-gray-600">
                                            <strong>Language:</strong> 
                                            {{ if (index $.LanguageIconMap (lower .Language)) }}
                                                <img src="{{ index $.LanguageIconMap (lower .Language) }}" alt="{{ .Language }}" class="w-4 h-4 mx-1">
                                            {{ end }}
                                            <span>{{ .Language }}</span>
                                        </div>
                                    </div>
                                    <ul class="flex flex-wrap gap-2 mt-2">
                                        {{ range .Occurrences }}
                                        <li class="text-xs text-gray-700 bg-gray-200 rounded-full px-3 py-1 border">{{ . }}</li>
                                        {{ end }}
                                    </ul>
                                </div>
                                {{ end }}
                            </div>
                        </td>
                    </tr>
                    {{ end }}
                </tbody>
            </table>
        </div>
    </div>
    </div>

    <script>
        const signatureSelect = document.getElementById('signatureSelect');
        const tagsSelect = document.getElementById('tagsSelect');
        const signatureChoices = new Choices(signatureSelect, { removeItemButton: true });
        const tagChoices = new Choices(tagsSelect, { removeItemButton: true });

        signatureSelect.addEventListener('change', filterTable);
        tagsSelect.addEventListener('change', filterTable);

        function filterTable() {
            const selectedSignatures = signatureChoices.getValue(true);
            const selectedTags = tagChoices.getValue(true);
            const rows = document.querySelectorAll('#signatureTable tr');

            for (let i = 0; i < rows.length; i += 2) {
                const row = rows[i];
                const rowSignature = row.getAttribute('data-signature');
                const rowTags = row.getAttribute('data-tags').split(',').map(tag => tag.trim());
                const detailRow = rows[i + 1];

                const signatureMatch = (selectedSignatures.length === 0 || selectedSignatures.includes(rowSignature));
                const tagsMatch = (selectedTags.length === 0 || selectedTags.some(tag => rowTags.includes(tag)));

                if (signatureMatch && tagsMatch) {
                    row.style.display = '';
                    detailRow.style.display = '';
                } else {
                    row.style.display = 'none';
                    detailRow.style.display = 'none';
                }
            }
        }

        function toggleDetails(id, btn) {
            var x = document.getElementById(id);
            var icon = btn.querySelector('.icon');
            if (x.classList.contains("hidden")) {
                x.classList.remove("hidden");
                icon.style.transform = "rotate(90deg)";
                btn.querySelector('span:last-child').textContent = "Hide Details";
            } else {
                x.classList.add("hidden");
                icon.style.transform = "rotate(0deg)";
                btn.querySelector('span:last-child').textContent = "Show Details";
            }
        }
    </script>
    <!-- Example Tailwind CSS Footer -->
    <footer class="bg-gray-50 text-gray-600 py-6 border-t border-gray-200">
      <div class="container mx-auto flex flex-col md:flex-row items-center justify-between px-4">
        <span class="text-sm">&copy; 2025 SafeDep. All rights reserved.</span>
        <div class="flex space-x-4 mt-4 md:mt-0">
          <a href="https://safedep.io/privacy" class="hover:text-gray-400 transition">Privacy Policy</a>
          <a href="https://safedep.io/terms" class="hover:text-gray-400 transition">Terms of Service</a>
        </div>
      </div>
    </footer>
    </body>
</html>`

	t := template.Must(template.New("report").Funcs(template.FuncMap{
		"splitTags": func(tags string) []string {
			tagList := strings.Split(tags, ",")
			for i := range tagList {
				tagList[i] = strings.TrimSpace(tagList[i])
			}
			return tagList
		},
		"lower": strings.ToLower,
	}).Parse(tmpl))

	f, err := os.Create(htmlPath)
	if err != nil {
		return err
	}
	defer f.Close()

	headers := []string{"Signature_ID", "Description", "Tags"}
	var rows []map[string]interface{}
	tagSet := make(map[string]struct{})
	for _, row := range hv.rows {
		rows = append(rows, map[string]interface{}{
			"Signature_ID":    row["Signature ID"],
			"Description":     row["Description"],
			"Tags":            row["Tags"],
			"FileOccurrences": row["FileOccurrences"],
		})

		tags := strings.Split(row["Tags"].(string), ",")
		for _, tag := range tags {
			tag = strings.TrimSpace(tag)
			tagSet[tag] = struct{}{}
		}
	}

	var uniqueTags []string
	for tag := range tagSet {
		uniqueTags = append(uniqueTags, tag)
	}

	return t.Execute(f, map[string]interface{}{
		"Headers":         headers,
		"Rows":            rows,
		"UniqueTags":      uniqueTags,
		"LanguageIconMap": languageIconMap,
	})
}

func VisualiseCodeAnalysisFindings(codeAnalysisFindings *codeanalysis.CodeAnalysisFindings, htmlPath string) error {
	hv := NewHTMLVisualiser([]string{"Signature ID", "Description", "Tags"})

	sigRows := map[string]map[string]interface{}{}

	for _, signatureResults := range codeAnalysisFindings.SignatureWiseMatchResults {
		// fmt.Println(sigid, ":", len(signatureResults), "matches")

		for _, match := range signatureResults {
			sig := match.MatchedSignature
			sigId := sig.Id
			desc := sig.Description
			tags := strings.Join(sig.Tags, ", ")

			fileMap := make(map[string]map[string]interface{})
			for _, condition := range match.MatchedConditions {
				for _, evidence := range condition.Evidences {
					evidenceDetailString := "Unknown"
					evidenceContent, exists := evidence.Metadata()
					if exists {
						evidenceDetailString = fmt.Sprintf("L%d:%d to L%d:%d",
							evidenceContent.StartLine, evidenceContent.StartColumn,
							evidenceContent.EndLine, evidenceContent.EndColumn)
					}
					conditionLocationString := fmt.Sprintf("%s - %s", condition.Condition.Type, strings.ReplaceAll(condition.Condition.Value, "\n", " "))
					key := match.FilePath + "|" + string(match.MatchedLanguageCode)
					if _, ok := fileMap[key]; !ok {
						fileMap[key] = map[string]interface{}{
							"File":        match.FilePath,
							"Language":    string(match.MatchedLanguageCode),
							"Occurrences": []string{},
						}
					}
					fileMap[key]["Occurrences"] = append(fileMap[key]["Occurrences"].([]string), fmt.Sprintf("%s - %s", conditionLocationString, evidenceDetailString))
				}
			}

			if _, ok := sigRows[sigId]; !ok {
				sigRows[sigId] = map[string]interface{}{
					"Signature ID":    sigId,
					"Description":     desc,
					"Tags":            tags,
					"FileOccurrences": []map[string]interface{}{},
				}
			}
			existing := sigRows[sigId]["FileOccurrences"].([]map[string]interface{})
			for _, v := range fileMap {
				existing = append(existing, v)
			}
			sigRows[sigId]["FileOccurrences"] = existing
		}
	}

	for _, row := range sigRows {
		hv.AddRow(row)
	}

	return hv.Finish(htmlPath)
}
