document.addEventListener('htmx:beforeRequest', (ev) => {
  // Before a request is made, clear out any error messages.
  ev.srcElement.parentElement.querySelector('#error').innerHTML = ''
})

document.addEventListener('htmx:afterRequest', (ev) => {
  const code = ev.detail.xhr.status
  if (code < 400) {
    // On a successful response, reset the form.
    ev.target?.reset && ev.target.reset()
  } else if (code === 401) {
    // TODO: make some sort of toast notificaiton
    alert("Unauthorized")
  } else if (code >= 500) {
    // TODO: make some sort of toast notificaiton
    alert("Something went wrong. Please try again later.")
  }
})
