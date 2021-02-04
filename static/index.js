function getEntries() {
    return axios({
        method: method,
        url: url,
        data: data,
        xsrfCookieName: 'csrftoken',
        xsrfHeaderName: 'X-CSRFToken',
        headers: {
            'X-Requested-With': 'XMLHttpRequest'
        }
    })
}

var app = new Vue({
    el: '#app',
    data: {
        key: 'Search',
        entries: [
            {code: 'abc', name: 'ABC', open: 123, high: 456, low: 0, close: 200}
        ]
    },
    created() {
        let hd = this;
        getEntries('', 'get').then((res) => {
            hd.entries = response.data.entries
        });
    },
    methods: {
        filterEntries() {
            let hd = this;
            let formData = new FormData();
            formData.append('title', this.search)
        }
    }
})