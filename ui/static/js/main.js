var navLinks = document.querySelectorAll("nav a")
for (var i = 0; i < navLinks.length; i++) {
  var link = navLinks[i]
  if (link.getAttribute("href") == window.location.pathname) {
    link.classList.add("live")
    break
  }
}

htmx.on("htmx:responseError", handleHtmxError)
htmx.on("htmx:sendError", handleHtmxError)

function handleHtmxError(ev) {
  const el = document.getElementById("error-notification")
  el.innerText = ev.detail.xhr.responseText
  el.classList.remove("hidden")
  setTimeout(() => el.classList.add("hidden"), 5_000)
}
