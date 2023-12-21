new Vue({
  el: '#eventList',
  data() {
    return {
      events: [],
      titleSearch: ""
    }
  },
  computed: {
    filteredEvents() {
      if (this.titleSearch == "") {
        return this.events
      } else {
        let filtered = [];
        for (let i = 0; i < this.events.length; i++) {
          if (this.events[i].Title.toLowerCase().startsWith(this.titleSearch)) {
            filtered.push(this.events[i])
          }
        }
        return filtered;
      }
    }
  },
  mounted() {
    let events = [];
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
          hour: '2-digit',
          minute: '2-digit'
        })
        const item = {
          Title: data[i].Title,
          Link: "/events/" + data[i].id,
          EventDate: d
        }
        events.push(item)
      }

      this.events = events;
    }
  },
  template: `
<div class="event-list">
  <h2>Events</h2>
  <div class="m-2">
    <form>
      <label for="title-filter">Search by name</label>
      <input id="title-filter" name="title-filter" v-model="titleSearch" type="text"></input>
    </form>
  </div>
  <ul class="event-list">
    <li v-for="ev in filteredEvents" class="event-list-item">
      <p><a v-bind:href="ev.Link">{{ev.Title}}</a> &#8212; {{ev.EventDate}}</p>
    </li>
  </ul>
</div>`
})

// Example eventList
// [
//   {
//     id: "2bdd0040-ac01-4127-a767-bdc681a15545",
//     "Title": "Event one",
//     "Description": "This is event one.",
//     "Date": "2023-11-14T02:35:55Z"
//   }
// ]