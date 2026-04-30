Write-Host "Packaging Helm chart..."
helm package ./charts/mlflow-k8s-operator -d ./docs

Write-Host "Updating Helm repository index..."
helm repo index ./docs --url https://NotHarshhaa.github.io/mlflow-k8s-operator

Write-Host "Chart packaged successfully!"
Write-Host ""
Write-Host "To publish to GitHub Pages:"
Write-Host "1. Switch to gh-pages branch: git checkout gh-pages"
Write-Host "2. Copy docs/* to root: Copy-Item -Path docs\* -Destination . -Recurse -Force"
Write-Host "3. Commit and push: git add .; git commit -m 'Update chart'; git push origin gh-pages"
Write-Host ""
Write-Host "Or use the GitHub Actions workflow by pushing to main branch."
