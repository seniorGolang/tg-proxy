(function () {
  function getPrefix() {
    var p = document.body.getAttribute("data-ui-prefix") || "";
    return p ? (p.startsWith("/") ? p : "/" + p) : "";
  }

  function getPathFromURL() {
    var prefix = getPrefix() || "/";
    var pathname = window.location.pathname;
    if (prefix === "/") {
      return pathname === "/" ? "" : pathname.slice(1);
    }
    if (pathname.startsWith(prefix)) {
      return pathname.slice(prefix.length).replace(/^\/+/, "");
    }
    return "";
  }

  var restore = { alias: null, version: null };

  function parsePath(path) {
    if (!path) return { alias: null, version: null };
    var parts = path.split("/").filter(Boolean);
    var alias = parts.length >= 1 ? parts[0] : null;
    var version = parts.length >= 2 ? parts[1] : null;
    return { alias: alias, version: version };
  }

  function getPkgFromSearch() {
    var q = new URLSearchParams(window.location.search);
    var pkg = q.get("pkg");
    return pkg != null && pkg !== "" ? pkg : null;
  }

  function getPackagesSkeletonHTML() {
    var rows = [];
    for (var i = 0; i < 5; i++) {
      rows.push(
        "<tr><td><div class=\"skeleton-item\"></div></td><td><div class=\"skeleton-item\"></div></td><td><div class=\"skeleton-item\"></div></td><td><div class=\"skeleton-item\"></div></td><td></td></tr>"
      );
    }
    return (
      '<div class="packages-view packages-skeleton">' +
      '<header class="manifest-header">' +
      '<div class="manifest-meta">' +
      '<div class="manifest-meta-row"><span class="manifest-meta-label">Источник</span><span class="skeleton-item" style="flex:1;max-width:16rem"></span></div>' +
      '<div class="manifest-meta-row"><span class="manifest-meta-label">Версия</span><span class="skeleton-item" style="width:4rem"></span></div>' +
      "</div>" +
      '<div class="manifest-command-block"><span class="skeleton-item"></span></div>' +
      "</header>" +
      '<table class="packages-table"><thead><tr><th>Пакет</th><th>Источник</th><th>Версия</th><th>Команда установки</th><th></th></tr></thead><tbody>' +
      rows.join("") +
      "</tbody></table></div>"
    );
  }

  function buildVersionPath(alias, version) {
    var prefix = getPrefix() || "/";
    if (prefix === "/") return "/" + alias + "/" + version;
    return prefix + "/" + alias + "/" + version;
  }

  function updateTreeActive() {
    var path = getPathFromURL();
    var parsed = parsePath(path);
    var alias = parsed.alias;
    var version = parsed.version;

    document.querySelectorAll(".tree-item.tree-active").forEach(function (el) {
      el.classList.remove("tree-active");
    });

    if (!alias) return;

    if (alias && version) {
      var versionItems = document.querySelectorAll("li.tree-item.version-item");
      for (var j = 0; j < versionItems.length; j++) {
        var link = versionItems[j].querySelector("[data-alias][data-version]");
        if (link && link.getAttribute("data-alias") === alias && link.getAttribute("data-version") === version) {
          versionItems[j].classList.add("tree-active");
          return;
        }
      }
      return;
    }

    var projectItems = document.querySelectorAll("li.tree-item.project-item");
    for (var k = 0; k < projectItems.length; k++) {
      var toggle = projectItems[k].querySelector(".tree-toggle");
      if (toggle && toggle.getAttribute("data-alias") === alias) {
        projectItems[k].classList.add("tree-active");
        return;
      }
    }
  }

  function copyToClipboard(text) {
    if (navigator.clipboard && navigator.clipboard.writeText) {
      navigator.clipboard.writeText(text);
    } else {
      var ta = document.createElement("textarea");
      ta.value = text;
      ta.setAttribute("readonly", "");
      ta.style.position = "fixed";
      ta.style.left = "-9999px";
      document.body.appendChild(ta);
      ta.select();
      try {
        document.execCommand("copy");
      } finally {
        document.body.removeChild(ta);
      }
    }
  }

  function doRestore(alias) {
    if (!alias) return;
    var prefix = getPrefix() || "/";
    var toggle = document.querySelector('.tree-toggle[data-alias="' + alias + '"]');
    if (!toggle) return;
    var li = toggle.closest("li.project-item");
    if (!li) return;
    var url = prefix + (prefix === "/" ? "" : "/") + "fragments/projects/" + alias + "/versions";
    htmx.ajax("GET", url, { target: li, swap: "outerHTML" });
  }

  function doRestorePackages(alias, version) {
    if (!alias || !version) return;
    var prefix = getPrefix() || "/";
    var url =
      prefix +
      (prefix === "/" ? "" : "/") +
      "fragments/projects/" +
      alias +
      "/versions/" +
      encodeURIComponent(version) +
      "/packages";
    htmx.ajax("GET", url, { target: "#content", swap: "innerHTML" });
  }

  function doRestorePackageCard(alias, version) {
    var q = new URLSearchParams(window.location.search);
    var pkg = q.get("pkg");
    if (pkg == null || pkg === "") return;
    if (!alias || !version) return;
    var prefix = getPrefix() || "/";
    var url =
      prefix +
      (prefix === "/" ? "" : "/") +
      "fragments/package?alias=" +
      encodeURIComponent(alias) +
      "&version=" +
      encodeURIComponent(version) +
      "&index=" +
      encodeURIComponent(pkg);
    htmx.ajax("GET", url, { target: "#content", swap: "innerHTML" });
  }

  function restoreFromURL() {
    var path = getPathFromURL();
    var parsed = parsePath(path);
    if (!parsed.alias) return;
    restore.alias = parsed.alias;
    restore.version = parsed.version;
    doRestore(parsed.alias);
  }

  function onAfterLoad(ev) {
    var path = (ev.detail && ev.detail.pathInfo && ev.detail.pathInfo.requestPath) || "";
    if (path.indexOf("fragments/projects") === -1) return;
    if (path.indexOf("/versions/") !== -1 && path.indexOf("/packages") !== -1) {
      if (restore.alias && restore.version) {
        doRestorePackageCard(restore.alias, restore.version);
        restore.alias = null;
        restore.version = null;
      }
      return;
    }
    if (path.indexOf("/versions") !== -1 && path.indexOf("/collapse") === -1) {
      if (restore.version) {
        setTimeout(function () {
          doRestorePackages(restore.alias, restore.version);
        }, 0);
      }
      return;
    }
    if (path.indexOf("fragments/projects") !== -1 && path.indexOf("/versions") === -1 && path.indexOf("/collapse") === -1) {
      setTimeout(function () {
        if (restore.alias) {
          doRestore(restore.alias);
        }
      }, 0);
    }
  }

  function onPopState() {
    var path = getPathFromURL();
    var parsed = parsePath(path);
    restore.alias = parsed.alias;
    restore.version = parsed.version;
    if (!parsed.alias) {
      document.getElementById("content").innerHTML = '<div class="welcome" id="welcome"><p>Выберите проект в дереве слева, затем версию и пакет.</p></div>';
      var expanded = document.querySelectorAll(".project-item.expanded");
      expanded.forEach(function (el) {
        var alias = el.querySelector(".tree-toggle").getAttribute("data-alias");
        if (alias) {
          var url = (getPrefix() || "/") + (getPrefix() === "/" ? "" : "/") + "fragments/projects/" + alias + "/collapse";
          htmx.ajax("GET", url, { target: el, swap: "outerHTML" });
        }
      });
      return;
    }
    doRestore(parsed.alias);
  }

  document.body.addEventListener("htmx:beforeRequest", function (ev) {
    var target = ev.detail && ev.detail.target;
    var elt = ev.detail && ev.detail.elt;
    if (!target || target.id !== "content" || !elt) return;
    var url = (elt.getAttribute && elt.getAttribute("hx-get")) || "";
    if (url.indexOf("/packages") === -1) return;
    var alias = elt.getAttribute("data-alias");
    var version = elt.getAttribute("data-version");
    if (alias && version) {
      history.pushState({ htmx: true }, "", buildVersionPath(alias, version));
      target.innerHTML = getPackagesSkeletonHTML();
      updateTreeActive();
    }
  });

  document.body.addEventListener("htmx:afterOnLoad", onAfterLoad);
  document.body.addEventListener("htmx:afterSwap", function (ev) {
    var target = ev.detail && ev.detail.target;
    if (target && target.id === "modal-content") {
      var modal = document.getElementById("modal");
      if (modal) {
        modal.classList.add("open");
        modal.setAttribute("aria-hidden", "false");
      }
    }
    updateTreeActive();
    if (ev.detail && ev.detail.xhr) {
      var pushUrl = ev.detail.xhr.getResponseHeader("HX-Push-Url");
      if (pushUrl && window.location.pathname !== pushUrl && window.location.pathname + window.location.search !== pushUrl) {
        history.pushState({ htmx: true }, "", pushUrl);
      }
    }
  });

  function closeModal() {
    var modal = document.getElementById("modal");
    if (modal) {
      modal.classList.remove("open");
      modal.setAttribute("aria-hidden", "true");
    }
  }

  document.body.addEventListener("click", function (ev) {
    if (ev.target.closest(".drawer-overlay") || ev.target.closest(".drawer-close")) {
      closeModal();
    }
  });

  document.addEventListener("keydown", function (ev) {
    if (ev.key === "Escape") {
      var modal = document.getElementById("modal");
      if (modal && modal.classList.contains("open")) closeModal();
    }
  });

  document.body.addEventListener("click", function (ev) {
    var btn = ev.target.closest(".copy-btn");
    if (btn) {
      var text = btn.getAttribute("data-copy");
      if (text) {
        ev.preventDefault();
        ev.stopPropagation();
        copyToClipboard(text);
        var originalContent = btn.textContent;
        btn.textContent = "\u2713";
        btn.classList.add("copied");
        if (btn._copyTimeout) clearTimeout(btn._copyTimeout);
        btn._copyTimeout = setTimeout(function () {
          btn.textContent = originalContent;
          btn.classList.remove("copied");
          btn._copyTimeout = null;
        }, 3000);
      }
    }
  });
  window.addEventListener("popstate", function (ev) {
    if (ev.state && ev.state.htmx) {
      onPopState();
    } else {
      onPopState();
    }
    updateTreeActive();
  });

  var initialPath = document.body.getAttribute("data-initial-path") || "";
  if (initialPath) {
    var parsed = parsePath(initialPath);
    restore.alias = parsed.alias;
    restore.version = parsed.version;
  } else {
    var urlPath = getPathFromURL();
    if (urlPath) {
      var parsed = parsePath(urlPath);
      restore.alias = parsed.alias;
      restore.version = parsed.version;
    }
  }

  updateTreeActive();
})();
