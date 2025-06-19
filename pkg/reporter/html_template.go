package reporter

var htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Signature Visualizer</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/choices.js/public/assets/styles/choices.min.css" />
    <script src="https://cdn.jsdelivr.net/npm/choices.js/public/assets/scripts/choices.min.js"></script>
</head>
<body>

<nav class="bg-white border-b border-gray-200 px-4 py-3 flex items-center justify-between">
    <a href="https://safedep.io/" class="flex items-center">
        <img src="https://avatars.githubusercontent.com/u/115209633?s=200&v=4" alt="Logo" class="h-10 w-10 rounded-md mr-3">
        <span class="font-semibold text-lg text-gray-800">SafeDep</span>
    </a>
    <a href="https://github.com/safedep/xbom" target="_blank" rel="noopener noreferrer"
       class="flex items-center bg-gray-100 hover:bg-gray-200 text-gray-800 font-medium px-4 py-2 rounded transition">
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
                    <tr>
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
                                    <div class="space-y-4 mt-2">
                                        {{ range $index, $item := .Matches }}
                                        <div class="border-t pt-4 first:border-t-0 first:pt-0">
                                            <div class="text-xs text-gray-700 bg-gray-200 rounded-full px-3 py-1 border inline-block">{{ $item.Occurrence }}</div>
                                            {{ if $item.Snippet }}
                                            <div class="mt-2">
                                                <div class="relative">
                                                    <button onclick="copyToClipboard(this)" class="absolute right-2 top-2 text-xs bg-gray-700 text-white px-2 py-1 rounded hover:bg-gray-600">
                                                        Copy
                                                    </button>
                                                    <pre class="bg-gray-900 text-[#1affeb] text-sm p-2 rounded overflow-x-auto font-mono"><code>{{ range $item.Snippet.Lines }}<span class="text-gray-500 mr-2">{{ .LineNum }}</span>{{ .Content }}
{{ end }}</code></pre>
                                                </div>
                                            </div>
                                            {{ end }}
                                        </div>
                                        {{ end }}
                                    </div>
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
function copyToClipboard(button) {
    const pre = button.nextElementSibling;
    const text = pre.textContent;
    navigator.clipboard.writeText(text).then(() => {
        const originalText = button.textContent;
        button.textContent = 'Copied!';
        button.disabled = true;
        setTimeout(() => {
            button.textContent = originalText;
            button.disabled = false;
        }, 2000);
    });
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
