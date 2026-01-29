(function () {
  const STORAGE_KEY = "webui-theme";
  const systemDark = window.matchMedia("(prefers-color-scheme: dark)");

  function getStored() {
    return localStorage.getItem(STORAGE_KEY);
  }

  function resolveTheme() {
    const stored = getStored();
    if (stored === "light" || stored === "dark") {
      return stored;
    }
    return systemDark.matches ? "dark" : "light";
  }

  function applyTheme(theme) {
    const root = document.documentElement;
    root.setAttribute("data-theme", theme);

    const btn = document.getElementById("theme-toggle");
    if (btn) {
      if (theme === "dark") {
        btn.classList.add("flip");
      } else {
        btn.classList.remove("flip");
      }
    }
  }

  function init() {
    applyTheme(resolveTheme());

    systemDark.addEventListener("change", function () {
      if (getStored() == null || getStored() === "") {
        applyTheme(resolveTheme());
      }
    });

    const btn = document.getElementById("theme-toggle");
    if (btn) {
      btn.addEventListener("click", function () {
        const current = resolveTheme();
        const next = current === "dark" ? "light" : "dark";
        localStorage.setItem(STORAGE_KEY, next);
        applyTheme(next);
      });
    }
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }
})();
