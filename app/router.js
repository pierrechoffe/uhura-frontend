import Ember from 'ember';
import config from './config/environment';

let Router = Ember.Router.extend({
  location: config.locationType
});

Router.map(function() {
  this.route('channel',       { path: '/channels/:channel_id' }, function() {
    //    this.route('episode',            { path: '/:episode_id' });
    this.modal('share-modal', { withParams: ['share'] });
  });

  //  Public route
  //  this.route('channel_by_url', { path: '/c/:channel_id' });
  //  this.route('search_by_url',  { path: '/s' });
  //  this.route('category',       { path: '/category/:category_id' });

  this.route('subscriptions');
  //  this.route('explore');
  this.route('privacy');
  this.route('login');
  this.route('register');
  this.route('settings', {});
});

export default Router;
