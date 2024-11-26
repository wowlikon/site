const buttonClasses = {
  Да: "green",
  Yes: "green",
  Нет: "red",
  No: "red",
};

document.querySelectorAll(".choice-button").forEach((button) => {
  const choice = button.getAttribute("data-choice");
  const className = buttonClasses[choice.trim()];

  if (className) {
    button.classList.add(className);
  }

  if (choice.startsWith(" ")) {
    button.addEventListener("mouseenter", () => moveButton(button));
    button.addEventListener("mouseover", () => moveButton(button));
    button.addEventListener("touchstart", () => moveButton(button));
  }
});

function moveButton(button) {
  const windowWidth = window.innerWidth;
  const windowHeight = window.innerHeight;

  const randomX = Math.random() * (windowWidth - button.offsetWidth);
  const randomY = Math.random() * (windowHeight - button.offsetHeight);

  button.style.position = "absolute";
  button.style.left = `${randomX}px`;
  button.style.top = `${randomY}px`;
}
