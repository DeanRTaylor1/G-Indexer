const queryInput = document.getElementById("queryInput");
const form = document.getElementById("queryForm");
const results = document.getElementById("results");
const resultSpeedBox = document.getElementById("resultSpeedBox");
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
  console.log(apiResult.Data);
  for (let result of apiResult.Data) {
    console.log(result);
    let newDiv = document.createElement("div");
    let url = document.createElement("a");
    url.href = result.path;
    url.innerText = result.path;
    url.target = "_blank";
    // let newContent = document.createTextNode(JSON.stringify(result.path));
    newDiv.appendChild(url);
    results.appendChild(newDiv);
  }
  resultSpeedBox.innerText = apiResult.Message;
}

document.addEventListener("DOMContentLoaded", async () => {
  const interval = setInterval(async () => {
    console.log("ping");
    const response = await fetch("/api/index-progress", {
      method: "GET",
      headers: {
        "Content-Type": "text/plain",
      },
    });
    const apiResult = await response.json();
    console.log(apiResult);

    if (apiResult.Data === 1) {
      clearInterval(interval);
    }
  }, 1000);
});
