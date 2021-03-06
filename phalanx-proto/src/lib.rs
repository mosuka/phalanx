pub mod phalanx {
    use std::fmt;

    use num::FromPrimitive;
    use serde::de::{self, Deserialize, Deserializer, MapAccess, SeqAccess, Visitor};
    use serde::ser::{Serialize, SerializeStruct, Serializer};

    tonic::include_proto!("phalanx");

    const NODE_DETAILS_FIELDS: &'static [&'static str] = &["address", "state", "role"];

    impl Serialize for NodeDetails {
        fn serialize<S>(&self, serializer: S) -> Result<S::Ok, S::Error>
        where
            S: Serializer,
        {
            let mut node_details =
                serializer.serialize_struct("NodeDetails", NODE_DETAILS_FIELDS.len())?;
            node_details.serialize_field("address", &self.address)?;
            node_details.serialize_field("state", &self.state)?;
            node_details.serialize_field("role", &self.role)?;
            node_details.end()
        }
    }

    impl<'de> Deserialize<'de> for NodeDetails {
        fn deserialize<D>(deserializer: D) -> Result<Self, D::Error>
        where
            D: Deserializer<'de>,
        {
            enum Field {
                Address,
                State,
                Role,
            }

            impl<'de> Deserialize<'de> for Field {
                fn deserialize<D>(deserializer: D) -> Result<Field, D::Error>
                where
                    D: Deserializer<'de>,
                {
                    struct FieldVisitor;

                    impl<'de> Visitor<'de> for FieldVisitor {
                        type Value = Field;

                        fn expecting(&self, formatter: &mut fmt::Formatter) -> fmt::Result {
                            formatter.write_str("`address`, `state` or `role`")
                        }

                        fn visit_str<E>(self, value: &str) -> Result<Field, E>
                        where
                            E: de::Error,
                        {
                            match value {
                                "address" => Ok(Field::Address),
                                "state" => Ok(Field::State),
                                "role" => Ok(Field::Role),
                                _ => Err(de::Error::unknown_field(value, NODE_DETAILS_FIELDS)),
                            }
                        }
                    }

                    deserializer.deserialize_identifier(FieldVisitor)
                }
            }

            struct NodeDetailsVisitor;

            impl<'de> Visitor<'de> for NodeDetailsVisitor {
                type Value = NodeDetails;

                fn expecting(&self, formatter: &mut fmt::Formatter) -> fmt::Result {
                    formatter.write_str("struct NodeDetails")
                }

                fn visit_seq<V>(self, mut seq: V) -> Result<NodeDetails, V::Error>
                where
                    V: SeqAccess<'de>,
                {
                    let address = seq
                        .next_element()?
                        .ok_or_else(|| de::Error::invalid_length(0, &self))?;
                    let state = seq
                        .next_element()?
                        .ok_or_else(|| de::Error::invalid_length(1, &self))?;
                    let role = seq
                        .next_element()?
                        .ok_or_else(|| de::Error::invalid_length(1, &self))?;

                    Ok(NodeDetails {
                        address,
                        state,
                        role,
                    })
                }

                fn visit_map<V>(self, mut map: V) -> Result<NodeDetails, V::Error>
                where
                    V: MapAccess<'de>,
                {
                    let mut address = None;
                    let mut state = None;
                    let mut role = None;
                    while let Some(key) = map.next_key()? {
                        match key {
                            Field::Address => {
                                if address.is_some() {
                                    return Err(de::Error::duplicate_field("address"));
                                }
                                address = Some(map.next_value()?);
                            }
                            Field::State => {
                                if state.is_some() {
                                    return Err(de::Error::duplicate_field("state"));
                                }
                                state = Some(map.next_value()?);
                            }
                            Field::Role => {
                                if role.is_some() {
                                    return Err(de::Error::duplicate_field("role"));
                                }
                                role = Some(map.next_value()?);
                            }
                        }
                    }
                    let address = address.ok_or_else(|| de::Error::missing_field("addres"))?;
                    let state = state.ok_or_else(|| de::Error::missing_field("state"))?;
                    let role = role.ok_or_else(|| de::Error::missing_field("role"))?;

                    Ok(NodeDetails {
                        address,
                        state,
                        role,
                    })
                }
            }

            deserializer.deserialize_struct("NodeDetails", NODE_DETAILS_FIELDS, NodeDetailsVisitor)
        }
    }

    impl FromPrimitive for State {
        fn from_i32(n: i32) -> Option<Self> {
            match n {
                _ if n == State::Disconnected as i32 => Some(State::Disconnected),
                _ if n == State::NotReady as i32 => Some(State::NotReady),
                _ if n == State::Ready as i32 => Some(State::Ready),
                _ => None,
            }
        }

        fn from_i64(n: i64) -> Option<Self> {
            match n {
                _ if n == State::Disconnected as i64 => Some(State::Disconnected),
                _ if n == State::NotReady as i64 => Some(State::NotReady),
                _ if n == State::Ready as i64 => Some(State::Ready),
                _ => None,
            }
        }

        fn from_u64(n: u64) -> Option<Self> {
            match n {
                _ if n == State::Disconnected as u64 => Some(State::Disconnected),
                _ if n == State::NotReady as u64 => Some(State::NotReady),
                _ if n == State::Ready as u64 => Some(State::Ready),
                _ => None,
            }
        }
    }

    impl FromPrimitive for Role {
        fn from_i32(n: i32) -> Option<Self> {
            match n {
                _ if n == Role::Candidate as i32 => Some(Role::Candidate),
                _ if n == Role::Replica as i32 => Some(Role::Replica),
                _ if n == Role::Primary as i32 => Some(Role::Primary),
                _ => None,
            }
        }

        fn from_i64(n: i64) -> Option<Self> {
            match n {
                _ if n == Role::Candidate as i64 => Some(Role::Candidate),
                _ if n == Role::Replica as i64 => Some(Role::Replica),
                _ if n == Role::Primary as i64 => Some(Role::Primary),
                _ => None,
            }
        }

        fn from_u64(n: u64) -> Option<Self> {
            match n {
                _ if n == Role::Candidate as u64 => Some(Role::Candidate),
                _ if n == Role::Replica as u64 => Some(Role::Replica),
                _ if n == Role::Primary as u64 => Some(Role::Primary),
                _ => None,
            }
        }
    }

    #[cfg(test)]
    mod tests {
        use crate::phalanx::{NodeDetails, Role, State};
        use num::FromPrimitive;

        #[test]
        fn test_deserialize() {
            let json_str = "{\"state\":0,\"role\":0,\"address\":\"0.0.0.0:5001\"}";
            let node_details: NodeDetails = serde_json::from_str(json_str).unwrap();

            assert_eq!(node_details.address, "0.0.0.0:5001");
            assert_eq!(node_details.state, State::Disconnected as i32);
            assert_eq!(node_details.role, Role::Candidate as i32);
        }

        #[test]
        fn test_serialize() {
            let node_details = NodeDetails {
                address: "0.0.0.0:5001".to_string(),
                state: State::Disconnected as i32,
                role: Role::Candidate as i32,
            };
            let json_str = serde_json::to_string(&node_details).unwrap();

            assert_eq!(
                json_str,
                "{\"address\":\"0.0.0.0:5001\",\"state\":0,\"role\":0}"
            );
        }

        #[test]
        fn test_state_from_i32() {
            let i: i32 = 1;
            let state = State::from_i32(i).unwrap();

            assert_eq!(state, State::NotReady);
        }

        #[test]
        fn test_state_from_i64() {
            let i: i64 = 2;
            let state = State::from_i64(i).unwrap();

            assert_eq!(state, State::Ready);
        }

        #[test]
        fn test_state_from_u64() {
            let i: u64 = 0;
            let state = State::from_u64(i).unwrap();

            assert_eq!(state, State::Disconnected);
        }

        #[test]
        fn test_role_from_i32() {
            let i: i32 = 1;
            let role = Role::from_i32(i).unwrap();

            assert_eq!(role, Role::Replica);
        }

        #[test]
        fn test_role_from_i64() {
            let i: i64 = 2;
            let role = Role::from_i64(i).unwrap();

            assert_eq!(role, Role::Primary);
        }

        #[test]
        fn test_role_from_u64() {
            let i: u64 = 0;
            let role = Role::from_u64(i).unwrap();

            assert_eq!(role, Role::Candidate);
        }
    }
}
