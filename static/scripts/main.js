document.addEventListener("htmx:beforeRequest", (ev) => {
  // Before a request is made, clear out any error messages.
  const el = ev.srcElement.parentElement.querySelector("#error")
  if (el) {
    el.innerHTML = ""
  }
})

document.addEventListener("htmx:afterRequest", (ev) => {
  const code = ev.detail.xhr.status
  if (code < 400) {
    // On a successful response, reset the form (if it was a form that triggered the request).
    ev.target?.reset && ev.target.reset()
  } else if (code === 401) {
    notify("Unauthorized")
  } else if (code === 429) {
    notify("Service under heavy load, come back later.")
  } else if (code >= 500) {
    notify("Something went wrong. Please try again later.")
  }
})

let timeoutId = null
function notify(msg) {
  if (timeoutId) clearTimeout(timeoutId)
  const el = document.querySelector("#notification")
  el.innerText = msg
  el.classList.remove("hideme")
  timeoutId = setTimeout(() => el.classList.add("hideme"), 3_000)
}
