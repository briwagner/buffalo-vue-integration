new Vue({
  el: '#eventForm',
  data() {
    return {
      options: [],
      form: {
        EventID: '',
        FullName: '',
        Email: '',
        authenticity_token: ''
      },
      formReturn: ''
    }
  },
  methods: {
    async submit() {
      this.formReturn = '';
      let formData = axios.toFormData(this.form);
      let resp = await axios({
        method: 'POST',
        url: '/app/add-guest',
        data: formData,
      }).catch((error) => {
        if (error) {
          if (error.response) {
            this.formReturn = 'Error ' + error.response.data;
            return
          } else {
            this.formReturn = 'Server error'
          }
        }
      })
      if (!resp) {
        return
      }

      this.formReturn = 'Reservation complete';
    }
  },
  mounted() {
    let options = [];
    const data = JSON.parse(eventList)
    if (!Array.isArray(data)) {
      console.log("not an array")
    }
    else {
      for (let i = 0; i < data.length; i++) {
        let d = new Date(data[i].Date).toLocaleDateString('en-us', {
          year: 'numeric',
          month: 'short',
          day: 'numeric',
        })
        const item = {
          ID: data[i].id,
          Label: data[i].Title + " (" + d + ")",
        }
        options.push(item)
      }

      this.options = options;
    }

    // Get auth token.
    let tok = document.querySelector('meta[name="csrf-token"]').content;
    this.form.authenticity_token = tok;
  },
  template: `
<div class="event-list">
  <p>Reserve your spot at one of the following events</p>
  <form id="reservationForm" action="/app/add-guest" method="POST" @submit.prevent="submit">
    <select name="EventID" id="EventID" v-model="form.EventID">
      <option value="">Select an event</option>
      <option v-for="(option) in options" :value="option.ID">{{option.Label}}</option>
    </select>
    <label for="FullName">Full name</label>
    <input type="text" name="FullName" id="FullName" v-model="form.FullName"></input>

    <label for="Email">Email</label>
    <input type="email" name="Email" id="Email" v-model="form.Email"></input>

    <input type="hidden" name="authenticity_token" v-model="form.authenticity_token">

    <button type="submit">Reserve a spot</button>
  </form>
  <div id="form-return">{{formReturn}}</div>
</div>`
})