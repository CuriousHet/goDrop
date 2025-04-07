document.addEventListener('DOMContentLoaded', function() {
    // Handle file upload
    const uploadForm = document.getElementById('uploadForm');
    if (uploadForm) {
        uploadForm.addEventListener('submit', function(e) {
            e.preventDefault();
            
            const formData = new FormData(this);
            const progressBar = document.getElementById('uploadProgress');
            const progressBarInner = progressBar.querySelector('.progress-bar');
            const uploadResult = document.getElementById('uploadResult');
            
            // Show progress bar
            progressBar.style.display = 'block';
            uploadResult.innerHTML = '';

            fetch('/upload', {
                method: 'POST',
                body: formData
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    uploadResult.innerHTML = `
                        <div class="alert alert-success">
                            File uploaded successfully!<br>
                            Hash: ${data.hash}<br>
                            <a href="/download?hash=${data.hash}" class="btn btn-sm btn-primary mt-2">Download</a>
                        </div>
                    `;
                    this.reset();
                } else {
                    uploadResult.innerHTML = `
                        <div class="alert alert-danger">
                            Upload failed: ${data.error}
                        </div>
                    `;
                }
            })
            .catch(error => {
                uploadResult.innerHTML = `
                    <div class="alert alert-danger">
                        Upload failed: ${error}
                    </div>
                `;
            })
            .finally(() => {
                progressBar.style.display = 'none';
            });
        });
    }

    // Handle file search
    const searchForm = document.getElementById('searchForm');
    if (searchForm) {
        searchForm.addEventListener('submit', function(e) {
            e.preventDefault();
            
            const searchQuery = this.querySelector('#search').value;
            const searchResults = document.getElementById('searchResults');
            
            if (!searchQuery) {
                searchResults.innerHTML = `
                    <div class="alert alert-warning">
                        Please enter a search term
                    </div>
                `;
                return;
            }

            searchResults.innerHTML = `
                <div class="text-center">
                    <div class="spinner-border text-primary" role="status">
                        <span class="visually-hidden">Loading...</span>
                    </div>
                </div>
            `;

            fetch(`/search?q=${encodeURIComponent(searchQuery)}`)
                .then(response => response.json())
                .then(data => {
                    if (data.results && data.results.length > 0) {
                        const resultsHtml = data.results.map(result => `
                            <div class="result-item">
                                <h6>${result.filename}</h6>
                                <p>Hash: ${result.hash}</p>
                                <p>Code Word: ${result.codeWord || 'None'}</p>
                                <a href="/download?hash=${result.hash}" class="btn btn-sm btn-primary">Download</a>
                            </div>
                        `).join('');
                        
                        searchResults.innerHTML = `
                            <div class="search-results">
                                ${resultsHtml}
                            </div>
                        `;
                    } else {
                        searchResults.innerHTML = `
                            <div class="alert alert-info">
                                No results found
                            </div>
                        `;
                    }
                })
                .catch(error => {
                    searchResults.innerHTML = `
                        <div class="alert alert-danger">
                            Search failed: ${error}
                        </div>
                    `;
                });
        });
    }
}); 