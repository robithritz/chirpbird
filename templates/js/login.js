async function submitLogin(evn) {
  evn.preventDefault();

  const username = document.getElementById('username');
  const password = document.getElementById('password');
  if (username.value == "" || password.value == "") {
    alert('please fill all the fields');
    return
  }

  resp = await postData(window.location.origin + "/login", {
    username: username.value,
    password: password.value
  })

  if (resp.status == false) {
    alert(resp.message);
    return
  }
  if (resp.token) {
    localStorage.setItem('token', resp.token);
    window.location.href = "/";
  }

}

async function postData(url, data) {
  const response = await fetch(url, {
    method: 'POST',
    cache: 'no-cache',
    headers: {
      'Content-Type': 'application/json'
      // 'Content-Type': 'application/x-www-form-urlencoded',
    },
    body: JSON.stringify(data)
  });
  return response.json()
}

const loginButton = document.getElementById('login-button');
const loginForm = document.getElementById('login-form')

loginButton.addEventListener('click', submitLogin);
loginForm.addEventListener('submit', function (evn) {
  evn.preventDefault();
  submitLogin(evn);
})