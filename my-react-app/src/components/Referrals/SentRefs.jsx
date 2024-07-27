import { useState, useEffect } from "react";
import axios from "axios";
import "../styles.css";

const SentRefs = () => {
  const [referrals, setReferrals] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchReferrals = async () => {
      try {
        const response = await axios.get(
          "http://localhost:8080/referrals-sent",
          { withCredentials: true }
        );
        setReferrals(response.data);
      } catch (error) {
        console.error("Error fetching referral requests:", error);
      } finally {
        setLoading(false);
      }
    };

    fetchReferrals();
  }, []);

  if (loading) return <p>Loading...</p>;

  return (
    <div>
      <h2>Referral Requests Sent</h2>
      {referrals.length === 0 ? (
        <p>No referral requests sent.</p>
      ) : (
        <ul>
          {referrals.map((referral) => (
            <li key={referral.id}>
              <strong>Referee Client:</strong> {referral.referee_client}
              <br />
              <strong>Title:</strong> {referral.title}
              <br />
              <strong>Context:</strong> {referral.content}
              <br />
              <strong>Referee Client Email:</strong>{" "}
              {referral.referee_client_email}
              <br />
              <strong>Referrer:</strong> {referral.referrer_username} :{" "}
              {referral.company_name}
              <br />
              {/* <strong>Company:</strong> {referral.company_name}
              <br /> */}
              <strong>Status:</strong> {referral.status}
              <br />
              <strong>Created At:</strong>{" "}
              {new Date(referral.created_at).toLocaleString()}
              <br />
              <hr />
            </li>
          ))}
        </ul>
      )}
    </div>
  );
};

export default SentRefs;
