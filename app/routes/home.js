import Ember from 'ember';

const { inject: { service }, RSVP: { hash } } = Ember;

export default Ember.Route.extend({
  client: service('uhura-client'),
  store: service('store'),

  model() {
    return hash({
      channels: this.get('client').request('', 'top', 'channels').then(this.fixImageURL),
      episodes: this.buildEpisodes()
    });
  },

  fixImageURL(data) {
    const { channels } = data;
    return channels.map((channel) => {
        channel.imageURL = channel.image_url;
        return channel;
      });
  },

  buildEpisodes() {
    return [
      {
        id: 86373,
        title: 'The Living Room',
        channel_id: 'loveradio',
        description: 'Diane\'s new neighbors across the way never shut their curtains, and that was the beginning of an intimate, but very one-sided relationship.'
      },
      {
        id: 465197,
        title: 'How to Become Batman',
        channel_id: 'invisibilia' ,
        description: 'In "How to Become Batman," Alix and Lulu examine the surprising effect that our expectations can have on the people around us. You\'ll hear how people\'s expectations can influence how well a rat runs a maze. Plus, the story of a man who is blind and says expectations have helped him see. Yes. See. This journey is not without skeptics.'
      }
    ];
  }
});