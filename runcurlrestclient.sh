echo "Post"
curl -X POST http://localhost:8081/v1/blog/create -d '{"blog": {"author_id": "GOD", "title": "- THE BLESSINGS", "content": "ARE INFINITE!"}}'

echo
echo "-------"
echo

echo "Get"
curl http://localhost:8081/v1/blog/60104b69d9616c87b396b3ff