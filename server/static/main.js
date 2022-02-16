const form = document.querySelector(".search-form");
form.addEventListener("submit", handleSubmit);

function handleSubmit(e) {
  e.preventDefault();
  let query = document.querySelector(".search-input").value;
  query = query.trim();
  getResults(query);
}

function getResults(query) {
  const endpoint = `/v1/indexes/enwiki/_search`;

  let headers = {
    'Content-Type': 'application/json',
  };

  let requestBody = {
    query: {
      type: 'boolean',
      options: {
        must: [
          {
            type: 'query_string',
            options: {
              query: `${query}`,
            },
          },
        ],
      },
    },
    start: 0,
    num: 10,
    sort_by: `-_score`,
    fields: [
      '*',
    ],
    highlights: {
      text: {
        highlighter: {
          type: 'html',
          options: {
            fragment_size: 200,
            pre_tag: '<mark>',
            post_tag: '</mark>',
          },
        },
        num: 3,
      },
    },
  };

  fetch(endpoint, {
    method: 'POST',
    headers: headers,
    body: JSON.stringify(requestBody),
  })
    .then((res) => res.json())
    .then((data) => {
      putResults(data);
    })
    .catch((e) => console.log(`ERROR : ${e}`));
}

function putResults(sResults) {
  console.log(sResults);
  const searchResults = document.querySelector(".results");
  searchResults.innerHTML = "";
  sResults.documents.forEach((doc) => {
    searchResults.insertAdjacentHTML(
      "beforeend",
      `<div class="result">
      <h3 class="result-title">
        <a href="${doc.fields.url[0]}" target="_blank" rel="noopener">${doc.fields.title[0]}</a>
      </h3>
      <span class="result-snippet">${doc.highlights.text}</span><br>
      <a href="${doc.fields.url[0]}" class="result-link" target="_blank" rel="noopener">${doc.fields.url[0]}</a>
    </div>`
    );
  });
}
