console.log("Shoo, Glar");

const queryInput = document.getElementById("query");
const form = document.getElementById("queryForm");
const results = document.getElementById("results");
form.addEventListener("submit", (event) => search(event, queryInput.value));
async function search(event, query) {
  event.preventDefault();
  console.log(query);
  results.innerHTML = "";
  const response = await fetch("/api/search", {
    method: "POST",
    headers: {
      "Content-Type": "text/plain",
    },
    body: query,
  });

  const apiResult = await response.json();
  console.log(apiResult);
  for (let result of apiResult) {
    console.log(result);
    let newDiv = document.createElement("div");
    let newContent = document.createTextNode(JSON.stringify(result.path));
    newDiv.appendChild(newContent);
    results.appendChild(newDiv);
  }
}
