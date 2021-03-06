pub mod etcd;
pub mod nop;

use std::fmt::Debug;
use std::io::Error as IOError;

use async_trait::async_trait;
use crossbeam::channel::Sender;
use dyn_clone::{clone_trait_object, DynClone};
use lazy_static::lazy_static;
use prometheus::{register_gauge_vec, GaugeVec};

pub const DEFAULT_ROOT: &str = "/phalanx";

lazy_static! {
    static ref NODE_STATE_GAUGE: GaugeVec = register_gauge_vec!(
        "phalanx_kvs_node_state",
        "Node state.",
        &["index", "shard", "node"]
    )
    .unwrap();
    static ref NODE_ROLE_GAUGE: GaugeVec = register_gauge_vec!(
        "phalanx_kvs_node_role",
        "Node role.",
        &["index", "shard", "node"]
    )
    .unwrap();
}

#[derive(Debug, Clone)]
pub struct KeyValuePair {
    pub key: String,
    pub value: Vec<u8>,
}

#[derive(Debug, Clone)]
pub enum EventType {
    Put,
    Delete,
}

#[derive(Debug, Clone)]
pub struct Event {
    pub event_type: EventType,
    pub key: String,
    pub value: Vec<u8>,
}

#[async_trait]
pub trait KeyValueStore: DynClone + Send + Sync + 'static {
    fn get_type(&self) -> &str;

    fn export_config_json(&self) -> Result<String, IOError>;

    async fn get(
        &mut self,
        key: &str,
    ) -> Result<Option<Vec<u8>>, Box<dyn std::error::Error + Send + Sync>>;

    async fn list(
        &mut self,
        prefix: &str,
    ) -> Result<Vec<KeyValuePair>, Box<dyn std::error::Error + Send + Sync>>;

    async fn put(
        &mut self,
        key: &str,
        value: Vec<u8>,
    ) -> Result<(), Box<dyn std::error::Error + Send + Sync>>;

    async fn delete(&mut self, key: &str) -> Result<(), Box<dyn std::error::Error + Send + Sync>>;

    async fn watch(
        &mut self,
        sender: Sender<Event>,
        key: &str,
    ) -> Result<(), Box<dyn std::error::Error + Send + Sync>>;

    async fn unwatch(&mut self) -> Result<(), Box<dyn std::error::Error + Send + Sync>>;
}

clone_trait_object!(KeyValueStore);

#[derive(Clone)]
pub struct KVSContainer {
    pub kvs: Box<dyn KeyValueStore>,
}
