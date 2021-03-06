import Ember from 'ember';
import MaterialDesignMixin from '../mixins/routes/material-design';

const { inject: { service }, set } = Ember;

export default Ember.Route.extend(MaterialDesignMixin, {
  session: service('session'),

  beforeModel(transition) {
    this.get('session').set('attemptedTransition', transition);
    return this._super(transition);
  },

  model(params) {
    return this.store.find('channel', params.channel_id).then((channel) => {
      // When the channel id is a feed url and the channel exists on
      // Uhura, the API returns the found channel, so the ID will be the
      // Uhura ID and not the feed url
      if (params.channel_id !== channel.id) {
        this.transitionTo('channel', channel.id);
      }
      return channel;
    });
  },

  setupController(controller, model) {
    this._super(controller, model);
    set(controller, 'episodesHasLink', !this.__isURL(model.id));
  },

  __isURL(id) {
    return id.match(/http:\/\//);
  }
});
