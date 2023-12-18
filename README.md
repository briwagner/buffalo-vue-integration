# Example app: Buffalo with VueJS integration

This is an example Buffalo app that integrates with VueJS components. The examples are intended to show two different ways to combine Buffalo and Vue.

These samples assume that Buffalo is managing all of the server-side operations. Both HTML pages and Javascript components are served by Buffalo, and the forms are handled by Buffalo.

Another option is to use a separate server to host the JS components, which fetch and send data to the Buffalo endpoints. That involves a more sophisticated setup with multiple servers, which is outside the scope of this example. A few modifications would allow this project to be split up (i.e. combine the Vue components into a single project) and hosted from separate machines.

## Event Planner app

<em>What does this app do?</em>

The app shows some basic CRUD patterns. Users can create an account with email and password. This uses a <a href="https://github.com/briwagner/buffalo-auth">forked version</a> of the Buffalo Auth plugin. Authenticated users can create Events. Any site visitor then can make reserve a spot at the event, using an email and name.

See the homepage for a list of routes.

## Javascript build step

This app does not use a frontend build process, in order to simplify the examples. The main `application.plush.html` template includes the script imports for Vue and Axios. Individual components are included directly on the relevant page templates.

For a production application, a little extra work should make it possible to add a bundling and/or build-process pipeline to these samples.

## Vue widgets

There are three Vue components shown here. Two of them read and display data. The third is a form component that submits data to the backend.

### 1. Search widget that reads page data

File: `public/assets/eventList.js` and `templates/events/all.plush.html`

Route: `/events`

This is one of the simplest ways to combine a server-generated web page along with Javascript components. It does not require the creation of additional server routes that exist solely to deliver JSON data to a frontend component. Another advantage is that the page, along with its data payload, can be cached to help improve page speed and minimize database operations.

The trick here is to use the server to render JSON data onto the page, where it is not visible to the user. The JS component then loads that JSON -- similar to how it would make a JSON request -- and renders the relevant data.

### 2. Search widget that loads

File: `public/assets/eventListRemote.js`

Route: `/events-remote`

Similar to #1, this sample loads data and provides a dynamic filter on the title field. Unlike the first, this one makes a call to the server to load the event list in the `mounted()` stage of the component.

### 3. Form that makes a remote server request

File: `public/assets/eventForm.js`

Route: `/app`

This form is a combination of both methods above. The server renders the event list to JSON and writes it to the page. The Vue component reads that data and generates a dynamic form. (Add frontend form validation and it'll do even more!) Finally, the user clicks submit, and Vue sends the form to the backend, showing the result of the operation on the page.