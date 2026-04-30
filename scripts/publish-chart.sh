#!/bin/bash
set -e

echo "Packaging Helm chart..."
helm package ./charts/mlflow-k8s-operator -d ./docs

echo "Updating Helm repository index..."
helm repo index ./docs --url https://NotHarshhaa.github.io/mlflow-k8s-operator

echo "Chart packaged successfully!"
echo ""
echo "To publish to GitHub Pages:"
echo "1. Switch to gh-pages branch: git checkout gh-pages"
echo "2. Copy docs/* to root: cp -r docs/* ."
echo "3. Commit and push: git add . && git commit -m 'Update chart' && git push origin gh-pages"
echo ""
echo "Or use the GitHub Actions workflow by pushing to main branch."
