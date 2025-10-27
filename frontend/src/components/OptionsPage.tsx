import { useNavigate } from "react-router";
import Options from "./Options";

export default function OptionsPage() {
  const navigate = useNavigate();

  return <Options onClose={() => navigate("/")} />;
}
