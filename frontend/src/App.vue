<script setup>
import { ref } from 'vue'

// HandleChopButton
// shortURL is the variable that will be used to store the short url

let shortURL = ref('')
const longURL = ref('')
const handleChopButton = async () => {
  shortURL.value = 'Loading'
  shortURL.value = await getShortCode(longURL.value)
  longURL.value = ''
}

// Method for sending post request to backend
const getShortCode = async (url) =>{
  const response = await fetch('https://chop-test.onrender.com/api/short', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      long: url
    })
  })
  const data = await response.json()
  return data
}
</script>

<template>
  <div id="mainbody" class="d-flex text-center text-white bg-dark">
  <div class="cover-container d-flex w-100 h-100 p-3 mx-auto flex-column">
  <header class="mb-5">
    <div>
      <h3 class="float-md-start mb-0">Chopper</h3>
      <nav class="nav nav-masthead justify-content-center float-md-end">
        <a class="nav-link active" aria-current="page" href="#">Home</a>
        <a class="nav-link" href="#">Features</a>
        <a class="nav-link" href="#">Contact</a>
      </nav>
    </div>
  </header>

  <main class="px-3 mt-5">
    <h1>Chopper</h1>
    <p class="lead">Shorten your links easily into small readble links.</p>

    <div class="input-group mb-3 lead">
      <input v-model="longURL" @keyup.enter="handleChopButton" type="text" class="form-control" placeholder="Enter URL" aria-label="Recipient's username" aria-describedby="button-addon2">
      <button @click="handleChopButton" class="btn btn-primary" type="button" id="button-addon2">Chop</button>
   </div>

  <!-- If shortURl is "Loading" show a spinner -->
  <div v-if="shortURL=='Loading'" class="spinner-border text-primary" role="status">
    <span class="visually-hidden">Loading...</span>
  </div>
  <div v-else-if="shortURL" class="alert alert-success w-50 mx-auto" role="alert">
    <h4 class="alert-heading">Your Short URL</h4>
    <!-- Dispplay the shortURL.short_url with a copy to clipbaord button attached. Keep btn as outline and input background a bit dark -->
    <div class="input-group mb-3">
      <input  type="text" class="form-control bg-dark text-white text-bold" :value="shortURL.short_url" aria-label="Recipient's username" aria-describedby="button-addon2">
      <button @click="navigator.clipboard.writeText(shortURL.short_url)" class="btn btn-sm btn-outline-dark" type="button" id="button-addon2">Copy</button>
    </div>

   
    <hr>
    <p class="mb-0">You can copy this link and share it with your friends.</p>
  </div>
  </main>

  <footer class="mt-auto text-white-50">
    <p>Made with Tears 🥹, by <a href="https://github.com/theakhandpatel" class="text-white">@theakhandpatel</a>.</p>
  </footer>
</div>
  </div>
</template>

<style>


#mainbody{
  height: 100vh;
}

* {
  padding: 0;
  margin: 0;
  box-sizing: border-box;
}

.cover-container {
  max-width: 60em;
}

</style>
