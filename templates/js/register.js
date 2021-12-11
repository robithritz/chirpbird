async function submitRegister(evn) {
  evn.preventDefault();
  const name = document.getElementById('name');
  const username = document.getElementById('username');
  const password = document.getElementById('password');
  const retypePassword = document.getElementById('retype-password');

  if (name.value == "" || username.value == "" || password.value == "") {
    alert('please fill all the fields');
    return
  }
  if (password.value !== retypePassword.value) {
    alert('retype password not match');
    return
  }

  resp = await postData(window.location.origin + "/users", {
    name: name.value,
    username: username.value,
    password: password.value
  })

  if (resp.status == false) {
    alert(resp.message);
    return
  } else {
    alert("Registation Succcesful.");
    window.location.href = "/login";

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

const registerButton = document.getElementById('register-button');
const registerForm = document.getElementById('register-form')

registerButton.addEventListener('click', submitRegister);
registerForm.addEventListener('submit', function (evn) {
  evn.preventDefault();
  submitRegister(evn);
})