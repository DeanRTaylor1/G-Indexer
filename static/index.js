const queryInput = document.getElementById("queryInput");
const queryForm = document.getElementById("queryForm");
const crawlForm = document.getElementById("crawlForm");
const crawlInput = document.getElementById("crawlInput");
const results = document.getElementById("results");
const progressBox = document.getElementById("progressBox");
const crawlSection = document.getElementById("crawlSection");
const querySection = document.getElementById("querySection");
const indexSelect = document.getElementById("indexSelect");
const statusBox = document.getElementById("statusBox");
const indexName = document.getElementById("indexName");
const resultsTitle = document.getElementById("resultsTitle");

indexSelect.addEventListener("change", (event) => startIndex(event));

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

function showQuerySection() {
  querySection.classList.remove("hidden");
  querySection.style.display = "flex";
}

function showResults() {
  resultsTitle.classList.remove("hidden");
  resultsTitle.style.display = "flex";
  results.style.display = "flex";
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
  console.log(apiResult);
  showResults();
  for (let result of apiResult.Data) {
    let newDiv = document.createElement("div");
    newDiv.classList.add("result-item");

    let title = document.createElement("a");
    title.href = result.path;

    title.classList.add("result-title");
    console.log(result.name);
    title.innerText = `${result.name.length === 1 ? result.path : result.name}`;
    title.target = "_blank";

    let url = document.createElement("div");
    url.classList.add("result-url");
    url.innerText = result.path;

    let description = document.createElement("div");
    description.classList.add("result-description");
    description.innerText = "";

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
      const stats = `Crawled ${apiResult.dir_length} pages, indexed ${apiResult.doc_count} pages, and found ${apiResult.term_count} terms.`;
      const link = document.createElement("a");
      const url = apiResult.index_name;

      // Check if the URL starts with 'http://' or 'https://'
      if (!url.startsWith("http://") && !url.startsWith("https://")) {
        link.href = "http://" + url;
      } else {
        link.href = url;
      }

      link.innerText = apiResult.index_name;
      link.className = "no-style-link";
      link.target = "_blank";
      indexName.innerHTML = "";
      indexName.appendChild(link);
      statusBox.innerText = "";
      statusBox.innerText = stats;
      if (
        apiResult.is_complete === true &&
        apiResult.message === "Not Started"
      ) {
        hideLoadingCircle();
        clearInterval(interval);
      } else if (apiResult.message === "In Progress") {
        hideLoadingCircle();
        showQuerySection();
      } else if (apiResult.message === "Not Started") {
        hideLoadingCircle();
        clearInterval(interval);
      } else if (apiResult.is_complete === true) {
        hideLoadingCircle();
        clearInterval(interval);
      }
    }, 300);
  } catch (error) {
    console.log(error);
    clearInterval(interval);
  }
};

const startCrawl = async (event, query) => {
  event.preventDefault();

  results.style.display = "none";
  showLoadingCircle();
  resultsTitle.style.display = "none";
  let timer;
  try {
    const response = await fetch("/api/crawl", {
      method: "POST",
      headers: {
        "Content-Type": "text/plain",
      },
      body: query,
    });
    const apiResult = await response.json();

    timer = setTimeout(() => {
      checkProgress();
      getIndexes();
    }, 2000);
  } catch (error) {
    console.log(error);
    clearTimeout(timer);
  }
};

const populateIndexSelect = (indexes) => {
  const selectElement = document.getElementById("indexSelect");
  // Clear any existing options
  selectElement.innerHTML = "";
  // Add a default option
  const defaultOption = document.createElement("option");
  defaultOption.text = "Select an index";
  selectElement.add(defaultOption);

  // Add options for each index
  indexes.forEach((index) => {
    const option = document.createElement("option");
    option.text = index;
    option.value = index;
    selectElement.add(option);
  });
};

const getIndexes = async () => {
  try {
    const response = await fetch("/api/indexes", {
      method: "GET",
      headers: {
        "Content-Type": "text/plain",
      },
    });
    const apiResult = await response.json();
    console.log(apiResult);
    populateIndexSelect(apiResult.Data);
  } catch (error) {
    console.log(error);
  }
};

const startIndex = async (event) => {
  event.preventDefault();
  results.style.display = "none";
  resultsTitle.style.display = "none";

  const index = indexSelect.value;
  console.log(index);
  if (index === "Select an index") {
    return;
  }
  let timer;
  try {
    const response = await fetch("/api/index", {
      method: "POST",
      headers: {
        "Content-Type": "text/plain",
      },
      body: index,
    });
    const apiResult = await response.json();
    console.log(apiResult);
    showLoadingCircle();
    timer = setTimeout(() => {
      checkProgress();
      getIndexes();
    }, 2000);
  } catch (error) {
    console.log(error);
    clearTimeout(timer);
  }
};

const startup = () => {
  showLoadingCircle();
  getIndexes();
  checkProgress();
};

document.addEventListener("DOMContentLoaded", startup);
