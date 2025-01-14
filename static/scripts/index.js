function isAuthenticated() {
  return !!localStorage.getItem("authToken");
}

function parseJwt(token) {
  const base64Url = token.split(".")[1];
  const base64 = base64Url.replace(/-/g, "+").replace(/_/g, "/");
  const jsonPayload = decodeURIComponent(
    atob(base64)
      .split("")
      .map(function (c) {
        return "%" + ("00" + c.charCodeAt(0).toString(16)).slice(-2);
      })
      .join(""),
  );

  return JSON.parse(jsonPayload);
}

function setupMenu() {
  const tabs = document.querySelector(".tabs");
  const avatar = document.getElementById("avatar");
  const dropdown = document.querySelector(".dropdown");
  const authForms = document.querySelector(".tab-content");
  const profileMenu = document.getElementById("profile-menu");
  const userName = document.getElementById("user-name");

  if (isAuthenticated()) {
    const token = localStorage.getItem("authToken");
    const userData = parseJwt(token);
    userName.textContent = userData.username || "User";
    avatar.src = `https://www.gravatar.com/avatar/${md5(userData.email)}?d=identicon`;

    // Скрываем элементы авторизации
    tabs.style.display = "none";
    authForms.style.display = "none";

    // Показываем профиль
    profileMenu.style.display = "block";

    // Закрываем выпадающее меню если оно открыто
    dropdown.style.display = "none";
  } else {
    avatar.src = "/static/images/default-avatar.png";

    // Показываем элементы авторизации
    tabs.style.display = "flex";
    authForms.style.display = "block";

    // Скрываем профиль
    profileMenu.style.display = "none";

    // Активируем первую вкладку по умолчанию
    const firstTab = document.querySelector(".tab-button");
    const firstPane = document.querySelector(".tab-pane");
    if (firstTab && firstPane) {
      firstTab.classList.add("active");
      firstPane.classList.add("active");
    }
  }
}

document.querySelectorAll(".tab-button").forEach((button) => {
  button.addEventListener("click", () => {
    document
      .querySelectorAll(".tab-button")
      .forEach((b) => b.classList.remove("active"));

    document
      .querySelectorAll(".tab-pane")
      .forEach((pane) => pane.classList.remove("active"));

    button.classList.add("active");

    const targetTab = document.getElementById(button.dataset.tab);
    targetTab.classList.add("active");
  });
});

document
  .getElementById("login-tab")
  .addEventListener("submit", async (event) => {
    event.preventDefault();
    const email = document.getElementById("login-email").value;
    const password = document.getElementById("login-password").value;

    try {
      const response = await fetch("/account/login", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ email, password }),
      });

      if (!response.ok) {
        const errorData = await response.json();
        alert(errorData.error);
        return;
      }

      const data = await response.json();
      localStorage.setItem("authToken", data.token);

      alert("Вход выполнен!");
      setupMenu();
    } catch (error) {
      console.error("Ошибка при входе:", error);
      alert("Произошла ошибка при входе.");
    }
  });

document
  .getElementById("register-tab")
  .addEventListener("submit", async (event) => {
    event.preventDefault();
    const email = document.getElementById("register-email").value;
    const username = document.getElementById("register-username").value;
    const password = document.getElementById("register-password").value;
    const confirmPassword = document.getElementById(
      "register-confirm-password",
    ).value;

    if (password !== confirmPassword) {
      alert("Пароли не совпадают!");
      return;
    }

    try {
      const response = await fetch("/account/register", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ username, email, password }),
      });

      if (!response.ok) {
        const errorData = await response.json();
        alert(errorData.error);
        return;
      }

      alert("Регистрация выполнена!");
    } catch (error) {
      console.error("Ошибка при регистрации:", error);
      alert("Произошла ошибка при регистрации.");
    }
  });

document.getElementById("logout-button").addEventListener("click", () => {
  localStorage.removeItem("authToken");

  alert("Вы вышли из аккаунта.");
  setupMenu();
});

document.querySelector(".menu-button").addEventListener("click", () => {
  const dropdown = document.querySelector(".dropdown");
  dropdown.style.display =
    dropdown.style.display === "block" ? "none" : "block";
});

function md5(string) {
  return CryptoJS.MD5(string).toString();
}

document.addEventListener("DOMContentLoaded", setupMenu);

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
