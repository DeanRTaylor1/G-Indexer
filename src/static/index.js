console.log("Shoo, Glar");

async function search() {
  const response = await fetch("/api/search", {
    method: "POST",
    headers: {
      "Content-Type": "text/plain",
    },
    body: "glsl function for linearly interpolation",
  });
  console.log(await response.json());
}

search();
