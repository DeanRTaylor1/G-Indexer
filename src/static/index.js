const queryInput = document.getElementById("queryInput");
const queryForm = document.getElementById("queryForm");
const crawlForm = document.getElementById("crawlForm");
const crawlInput = document.getElementById("crawlInput");
const results = document.getElementById("results");
const progressBox = document.getElementById("progressBox");
const crawlSection = document.getElementById("crawlSection");
const querySection = document.getElementById("querySection");

queryForm.addEventListener("submit", (event) =>
  search(event, queryInput.value)
);
crawlForm.addEventListener("submit", (event) =>
  startCrawl(event, crawlInput.value)
);

function showLoadingCircle() {
  const loadingCircle = document.getElementById("loading-circle");
  loadingCircle.classList.remove("hidden");
}

function hideLoadingCircle() {
  const loadingCircle = document.getElementById("loading-circle");
  loadingCircle.classList.add("hidden");
}

function showCrawlSection() {
  crawlSection.classList.remove("hidden");
  crawlSection.style.display = "flex";
}

function hideCrawlSection() {
  crawlSection.classList.add("hidden");
  crawlSection.style.display = "hidden";
}

function showQuerySection() {
  querySection.classList.remove("hidden");
  querySection.style.display = "flex";
}

function showResultsTitle() {
  const resultsTitle = document.getElementById("resultsTitle");
  resultsTitle.classList.remove("hidden");
  resultsTitle.style.display = "flex";
}

async function search(event, query) {
  event.preventDefault();
  console.log(query);
  results.innerHTML = "";
  progressBox.innerText = "";

  const response = await fetch("/api/search", {
    method: "POST",
    headers: {
      "Content-Type": "text/plain",
    },
    body: query,
  });

  const apiResult = await response.json();

  showResultsTitle();
  for (let result of apiResult.Data) {
    let newDiv = document.createElement("div");
    newDiv.classList.add("result-item");

    let title = document.createElement("a");
    title.href = result.path;
    let pathParts = result.name.split(" > ");

    if (pathParts[pathParts.length - 1].trim() === "") {
      pathParts.pop();
    }

    let formattedPath = `${pathParts.slice(1).join(" > ")}`;
    formattedPath[-1] === "";
    title.classList.add("result-title");
    title.innerText = `${
      result.name.length === 1 ? result.path : formattedPath
    }`;
    title.target = "_blank";

    let url = document.createElement("div");
    url.classList.add("result-url");
    url.innerText = result.path;

    let description = document.createElement("div");
    description.classList.add("result-description");
    description.innerText =
      "This is a short description for the result. Replace this with the actual description from your API data.";

    newDiv.appendChild(title);
    newDiv.appendChild(url);
    newDiv.appendChild(description);
    results.appendChild(newDiv);
  }
  progressBox.innerText = apiResult.Message;
}

const checkProgress = async () => {
  try {
    let apiResult;
    const interval = setInterval(async () => {
      console.log("ping");
      const response = await fetch("/api/progress", {
        method: "GET",
        headers: {
          "Content-Type": "text/plain",
        },
      });
      apiResult = await response.json();
      console.log(apiResult);

      if (
        apiResult.is_complete === true &&
        apiResult.message === "Not Started"
      ) {
        hideLoadingCircle();
        showCrawlSection();
        clearInterval(interval);
      } else if (
        apiResult.is_complete === true ||
        apiResult.message === "In Progress"
      ) {
        hideLoadingCircle();
        showQuerySection();
        clearInterval(interval);
      }
    }, 300);
  } catch (error) {
    console.log(error);
    clearInterval(interval);
  }
};

const startCrawl = async (event, query) => {
  hideCrawlSection();
  showLoadingCircle();
  event.preventDefault();
  try {
    const response = await fetch("/api/crawl", {
      method: "POST",
      headers: {
        "Content-Type": "text/plain",
      },
      body: query,
    });
    const apiResult = await response.json();

    const timer = setTimeout(() => {
      checkProgress();
    }, 750);
  } catch (error) {
    console.log(error);
    clearTimeout(timer);
  }
};

const startup = () => {
  showLoadingCircle();
  checkProgress();
};

document.addEventListener("DOMContentLoaded", startup);
