document.addEventListener("DOMContentLoaded", () => {
  const loginForm = document.getElementById("login-form");
  const registerForm = document.getElementById("register-form");
  const avatar = document.getElementById("user-avatar");
  const dropdownMenu = document.getElementById("dropdown-menu");

  // Пример обработки логина
  loginForm.addEventListener("submit", async (event) => {
    event.preventDefault();

    const formData = new FormData(loginForm);
    const response = await fetch("/login", {
      method: "POST",
      body: formData,
    });

    if (response.ok) {
      const data = await response.json();
      avatar.src = `https://www.gravatar.com/avatar/${data.gravatarHash}?s=80&d=identicon`;
      dropdownMenu.innerHTML = `<p>Welcome, ${data.username}!</p>`;
    } else {
      alert("Login failed");
    }
  });

  // Пример обработки регистрации
  registerForm.addEventListener("submit", async (event) => {
    event.preventDefault();

    const formData = new FormData(registerForm);
    const response = await fetch("/register", {
      method: "POST",
      body: formData,
    });

    if (response.ok) {
      alert("Registration successful! Please log in.");
      registerForm.reset();
    } else {
      alert("Registration failed");
    }
  });
});

let languageColors = {};

fetch(
  "https://raw.githubusercontent.com/ozh/github-colors/refs/heads/master/colors.json",
)
  .then((response) => response.json())
  .then((data) => {
    languageColors = data;
    const widgets = document.querySelectorAll(".github-widget");

    widgets.forEach((widget) => {
      const username = widget.getAttribute("data-username");
      const repo = widget.getAttribute("data-repo");

      fetch(`/api/repos/${username}/${repo}`)
        .then((response) => {
          if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
          }
          return response.json();
        })
        .then((data) => {
          const repoNameElement = widget.querySelector("h2");
          const repoLinkElement = document.createElement("a");

          // Доступ к данным о репозитории
          repoLinkElement.href = data.repository.html_url;
          repoLinkElement.target = "_blank";
          repoLinkElement.innerText = data.repository.name || "Нет названия";

          repoNameElement.innerHTML = "";
          repoNameElement.appendChild(repoLinkElement);
          widget.querySelector("p").innerText =
            data.repository.description || "Нет описания";
          widget.querySelector("span").innerText =
            data.repository.stargazers_count || 0;

          createLanguageBar(widget, data.languages);
          createLanguageLegend(widget, data.languages);
        })
        .catch((error) => {
          console.error("Error loading repo data:", error);
          widget.querySelector("h2").innerText = "Ошибка загрузки данных";
          widget.querySelector("p").innerText = error.message;
        });
    });
  })
  .catch((error) => console.error("Error loading colors:", error));

function createLanguageBar(widget, languages) {
  const languageBar = widget.querySelector(".language-bar");
  languageBar.innerHTML = "";
  const totalBytes = Object.values(languages).reduce((a, b) => a + b, 0);

  let otherBytes = 0;
  const displayedLanguages = {};

  for (const [language, bytes] of Object.entries(languages)) {
    const percentage = (bytes / totalBytes) * 100;

    if (percentage >= 1) {
      displayedLanguages[language] = percentage;
    } else {
      otherBytes += bytes;
    }
  }

  for (const [language, percentage] of Object.entries(displayedLanguages)) {
    const segment = document.createElement("div");
    segment.style.width = `${percentage}%`;
    segment.style.backgroundColor = getLanguageColor(language);
    segment.classList.add("language-segment");
    languageBar.appendChild(segment);
  }

  if (otherBytes > 0) {
    const otherPercentage = (otherBytes / totalBytes) * 100;
    const otherSegment = document.createElement("div");
    otherSegment.style.width = `${otherPercentage}%`;
    otherSegment.classList.add("language-segment");
    otherSegment.style.backgroundColor = "#cccccc";
    languageBar.appendChild(otherSegment);
  }
}

function createLanguageLegend(widget, languages) {
  const legend = widget.querySelector(".language-legend");
  legend.innerHTML = "";

  let otherCount = 0;

  for (const [language, bytes] of Object.entries(languages)) {
    const percentage =
      (bytes / Object.values(languages).reduce((a, b) => a + b, 0)) * 100;

    if (percentage >= 1) {
      const legendItem = document.createElement("div");
      legendItem.classList.add("legend-item");

      const colorBox = document.createElement("span");
      colorBox.style.backgroundColor = getLanguageColor(language);
      colorBox.classList.add("color-box");

      const text = document.createElement("span");
      text.innerText = `${language}: ${percentage.toFixed(2)}%`;

      legendItem.appendChild(colorBox);
      legendItem.appendChild(text);
      legend.appendChild(legendItem);
    } else {
      otherCount += bytes;
    }
  }

  if (otherCount > 0) {
    const otherLegendItem = document.createElement("div");
    otherLegendItem.classList.add("legend-item");

    const colorBox = document.createElement("span");
    colorBox.style.backgroundColor = "#cccccc";
    colorBox.classList.add("color-box");

    const text = document.createElement("span");
    const otherPercentage =
      (otherCount / Object.values(languages).reduce((a, b) => a + b, 0)) * 100;
    text.innerText = `Other: ${otherPercentage.toFixed(2)}%`;

    otherLegendItem.appendChild(colorBox);
    otherLegendItem.appendChild(text);
    legend.appendChild(otherLegendItem);
  }
}

function getLanguageColor(language) {
  return languageColors[language].color || "#cccccc";
}

let currentImageIndex = {};

function showGallery(language, headerElement) {
  const carousels = document.querySelectorAll(".carousel");
  carousels.forEach((carousel) => {
    carousel.classList.remove("active");
  });

  const activeCarousel = document.getElementById(language);
  activeCarousel.classList.add("active");

  const headers = document.querySelectorAll(".sidebar h2");
  headers.forEach((header) => {
    header.classList.remove("active");
  });

  headerElement.classList.add("active");

  currentImageIndex[language] = 0;
  updateImages(language);
  updateIndicators(language);
}

function changeImage(galleryId, direction) {
  const images = document.querySelectorAll(`#${galleryId} img`);
  const totalImages = images.length;

  currentImageIndex[galleryId] =
    (currentImageIndex[galleryId] + direction + totalImages) % totalImages;
  updateImages(galleryId);
  updateIndicators(galleryId);
}

function updateImages(galleryId) {
  const images = document.querySelectorAll(`#${galleryId} img`);
  images.forEach((img, index) => {
    img.style.display =
      index === currentImageIndex[galleryId] ? "block" : "none";
  });
}

function updateIndicators(galleryId) {
  const indicatorsContainer = document.querySelector(
    `#${galleryId} .indicator`,
  );
  indicatorsContainer.innerHTML = "";
  const totalImages = document.querySelectorAll(`#${galleryId} img`).length;

  for (let i = 0; i < totalImages; i++) {
    const dot = document.createElement("span");
    dot.className = i === currentImageIndex[galleryId] ? "active" : "";
    dot.onclick = () => {
      currentImageIndex[galleryId] = i;
      updateImages(galleryId);
      updateIndicators(galleryId);
    };
    indicatorsContainer.appendChild(dot);
  }
}

showGallery("python", document.body.getElementsByClassName("active")[0]);
